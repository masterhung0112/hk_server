package storetest

import (
	"fmt"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/store"
	"github.com/masterhung0112/hk_server/v5/store/searchtest"
	"github.com/masterhung0112/hk_server/v5/store/sqlstore"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/mattn/go-sqlite3"
)

type SqlStore interface {
	GetMaster() *gorp.DbMap
	DriverName() string
}

type StoreType struct {
	Name        string
	SqlSettings *model.SqlSettings
	SqlStore    SqlStore
	Store       store.Store
}

var StoreTypes []*StoreType = []*StoreType{}

func newStoreType(name, driver string) *StoreType {
	return &StoreType{
		Name:        name,
		SqlSettings: MakeSqlSettings(driver, false),
	}
}

func StoreTest(t *testing.T, f func(*testing.T, store.Store)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range StoreTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) {
			if testing.Short() {
				t.SkipNow()
			}
			f(t, st.Store)
		})
	}
}

func StoreTestWithSearchTestEngine(t *testing.T, f func(*testing.T, store.Store, *searchtest.SearchTestEngine)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()

	for _, st := range StoreTypes {
		st := st
		searchTestEngine := &searchtest.SearchTestEngine{
			Driver: *st.SqlSettings.DriverName,
		}

		t.Run(st.Name, func(t *testing.T) { f(t, st.Store, searchTestEngine) })
	}
}

func StoreTestWithSqlStore(t *testing.T, f func(*testing.T, store.Store, SqlStore)) {
	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	for _, st := range StoreTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) {
			if testing.Short() {
				t.SkipNow()
			}
			f(t, st.Store, st.SqlStore)
		})
	}
}

func initStores() {
	if testing.Short() {
		return
	}
	// In CI, we already run the entire test suite for both mysql and postgres in parallel.
	// So we just run the tests for the current database set.
	if os.Getenv("IS_CI") == "true" {
		switch os.Getenv("MM_SQLSETTINGS_DRIVERNAME") {
		case "mysql":
			StoreTypes = append(StoreTypes, newStoreType("MySQL", model.DATABASE_DRIVER_MYSQL))
		case "postgres":
			StoreTypes = append(StoreTypes, newStoreType("PostgreSQL", model.DATABASE_DRIVER_POSTGRES))
		}
	} else {
		StoreTypes = append(StoreTypes, newStoreType("MySQL", model.DATABASE_DRIVER_MYSQL),
			newStoreType("PostgreSQL", model.DATABASE_DRIVER_POSTGRES))
	}

	defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
	}()
	var wg sync.WaitGroup
	for _, st := range StoreTypes {
		st := st
		wg.Add(1)
		go func() {
			defer wg.Done()
			sqlStore := sqlstore.New(*st.SqlSettings, nil)
			st.Store = sqlStore
			st.SqlStore = sqlStore
			st.Store.DropAllTables()
			st.Store.MarkSystemRanUnitTests()
		}()
	}
	wg.Wait()
}

var tearDownStoresOnce sync.Once

func tearDownStores() {
	if testing.Short() {
		return
	}
	tearDownStoresOnce.Do(func() {
		var wg sync.WaitGroup
		wg.Add(len(StoreTypes))
		for _, st := range StoreTypes {
			st := st
			go func() {
				if st.Store != nil {
					st.Store.Close()
				}
				if st.SqlSettings != nil {
					CleanupSqlSettings(st.SqlSettings)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

// This test was used to consistently reproduce the race
// before the fix in MM-28397.
// Keeping it here to help avoiding future regressions.
func TestStoreLicenseRace(t *testing.T) {
	settings := makeSqlSettings(model.DATABASE_DRIVER_POSTGRES)
	store := sqlstore.New(*settings, nil)

	wg := sync.WaitGroup{}
	wg.Add(3)

	go func() {
		store.UpdateLicense(&model.License{})
		wg.Done()
	}()

	go func() {
		store.GetReplica()
		wg.Done()
	}()

	go func() {
		store.GetSearchReplica()
		wg.Done()
	}()

	wg.Wait()
}

func TestGetReplica(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description                string
		DataSourceReplicaNum       int
		DataSourceSearchReplicaNum int
	}{
		{
			"no replicas",
			0,
			0,
		},
		{
			"one source replica",
			1,
			0,
		},
		{
			"multiple source replicas",
			3,
			0,
		},
		{
			"one source search replica",
			0,
			1,
		},
		{
			"multiple source search replicas",
			0,
			3,
		},
		{
			"one source replica, one source search replica",
			1,
			1,
		},
		{
			"one source replica, multiple source search replicas",
			1,
			3,
		},
		{
			"multiple source replica, one source search replica",
			3,
			1,
		},
		{
			"multiple source replica, multiple source search replicas",
			3,
			3,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description+" with license", func(t *testing.T) {
			t.Parallel()

			settings := makeSqlSettings(model.DATABASE_DRIVER_POSTGRES)
			dataSourceReplicas := []string{}
			dataSourceSearchReplicas := []string{}
			for i := 0; i < testCase.DataSourceReplicaNum; i++ {
				dataSourceReplicas = append(dataSourceReplicas, *settings.DataSource)
			}
			for i := 0; i < testCase.DataSourceSearchReplicaNum; i++ {
				dataSourceSearchReplicas = append(dataSourceSearchReplicas, *settings.DataSource)
			}

			settings.DataSourceReplicas = dataSourceReplicas
			settings.DataSourceSearchReplicas = dataSourceSearchReplicas
			store := sqlstore.New(*settings, nil)
			defer func() {
				store.Close()
				CleanupSqlSettings(settings)
			}()

			store.UpdateLicense(&model.License{})

			replicas := make(map[*gorp.DbMap]bool)
			for i := 0; i < 5; i++ {
				replicas[store.GetReplica()] = true
			}

			searchReplicas := make(map[*gorp.DbMap]bool)
			for i := 0; i < 5; i++ {
				searchReplicas[store.GetSearchReplica()] = true
			}

			if testCase.DataSourceReplicaNum > 0 {
				// If replicas were defined, ensure none are the master.
				assert.Len(t, replicas, testCase.DataSourceReplicaNum)

				for replica := range replicas {
					assert.NotSame(t, store.GetMaster(), replica)
				}

			} else if assert.Len(t, replicas, 1) {
				// Otherwise ensure the replicas contains only the master.
				for replica := range replicas {
					assert.Same(t, store.GetMaster(), replica)
				}
			}

			if testCase.DataSourceSearchReplicaNum > 0 {
				// If search replicas were defined, ensure none are the master nor the replicas.
				assert.Len(t, searchReplicas, testCase.DataSourceSearchReplicaNum)

				for searchReplica := range searchReplicas {
					assert.NotSame(t, store.GetMaster(), searchReplica)
					for replica := range replicas {
						assert.NotSame(t, searchReplica, replica)
					}
				}
			} else if testCase.DataSourceReplicaNum > 0 {
				assert.Equal(t, len(replicas), len(searchReplicas))
				for k := range replicas {
					assert.True(t, searchReplicas[k])
				}
			} else if testCase.DataSourceReplicaNum == 0 && assert.Len(t, searchReplicas, 1) {
				// Otherwise ensure the search replicas contains the master.
				for searchReplica := range searchReplicas {
					assert.Same(t, store.GetMaster(), searchReplica)
				}
			}
		})

		t.Run(testCase.Description+" without license", func(t *testing.T) {
			t.Parallel()

			settings := makeSqlSettings(model.DATABASE_DRIVER_POSTGRES)
			dataSourceReplicas := []string{}
			dataSourceSearchReplicas := []string{}
			for i := 0; i < testCase.DataSourceReplicaNum; i++ {
				dataSourceReplicas = append(dataSourceReplicas, *settings.DataSource)
			}
			for i := 0; i < testCase.DataSourceSearchReplicaNum; i++ {
				dataSourceSearchReplicas = append(dataSourceSearchReplicas, *settings.DataSource)
			}

			settings.DataSourceReplicas = dataSourceReplicas
			settings.DataSourceSearchReplicas = dataSourceSearchReplicas
			store := sqlstore.New(*settings, nil)
			defer func() {
				store.Close()
				CleanupSqlSettings(settings)
			}()

			replicas := make(map[*gorp.DbMap]bool)
			for i := 0; i < 5; i++ {
				replicas[store.GetReplica()] = true
			}

			searchReplicas := make(map[*gorp.DbMap]bool)
			for i := 0; i < 5; i++ {
				searchReplicas[store.GetSearchReplica()] = true
			}

			if testCase.DataSourceReplicaNum > 0 {
				// If replicas were defined, ensure none are the master.
				assert.Len(t, replicas, 1)

				for replica := range replicas {
					assert.Same(t, store.GetMaster(), replica)
				}

			} else if assert.Len(t, replicas, 1) {
				// Otherwise ensure the replicas contains only the master.
				for replica := range replicas {
					assert.Same(t, store.GetMaster(), replica)
				}
			}

			if testCase.DataSourceSearchReplicaNum > 0 {
				// If search replicas were defined, ensure none are the master nor the replicas.
				assert.Len(t, searchReplicas, 1)

				for searchReplica := range searchReplicas {
					assert.Same(t, store.GetMaster(), searchReplica)
				}

			} else if testCase.DataSourceReplicaNum > 0 {
				assert.Equal(t, len(replicas), len(searchReplicas))
				for k := range replicas {
					assert.True(t, searchReplicas[k])
				}
			} else if assert.Len(t, searchReplicas, 1) {
				// Otherwise ensure the search replicas contains the master.
				for searchReplica := range searchReplicas {
					assert.Same(t, store.GetMaster(), searchReplica)
				}
			}
		})
	}
}

func TestGetDbVersion(t *testing.T) {
	testDrivers := []string{
		model.DATABASE_DRIVER_POSTGRES,
		model.DATABASE_DRIVER_MYSQL,
	}

	for _, driver := range testDrivers {
		t.Run("Should return db version for "+driver, func(t *testing.T) {
			t.Parallel()
			settings := makeSqlSettings(driver)
			store := sqlstore.New(*settings, nil)

			version, err := store.GetDbVersion(false)
			require.Nil(t, err)
			require.Regexp(t, regexp.MustCompile(`\d+\.\d+(\.\d+)?`), version)
		})
	}
}

func TestGetAllConns(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		Description                string
		DataSourceReplicaNum       int
		DataSourceSearchReplicaNum int
		ExpectedNumConnections     int
	}{
		{
			"no replicas",
			0,
			0,
			1,
		},
		{
			"one source replica",
			1,
			0,
			2,
		},
		{
			"multiple source replicas",
			3,
			0,
			4,
		},
		{
			"one source search replica",
			0,
			1,
			1,
		},
		{
			"multiple source search replicas",
			0,
			3,
			1,
		},
		{
			"one source replica, one source search replica",
			1,
			1,
			2,
		},
		{
			"one source replica, multiple source search replicas",
			1,
			3,
			2,
		},
		{
			"multiple source replica, one source search replica",
			3,
			1,
			4,
		},
		{
			"multiple source replica, multiple source search replicas",
			3,
			3,
			4,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Description, func(t *testing.T) {
			t.Parallel()
			settings := makeSqlSettings(model.DATABASE_DRIVER_POSTGRES)
			dataSourceReplicas := []string{}
			dataSourceSearchReplicas := []string{}
			for i := 0; i < testCase.DataSourceReplicaNum; i++ {
				dataSourceReplicas = append(dataSourceReplicas, *settings.DataSource)
			}
			for i := 0; i < testCase.DataSourceSearchReplicaNum; i++ {
				dataSourceSearchReplicas = append(dataSourceSearchReplicas, *settings.DataSource)
			}

			settings.DataSourceReplicas = dataSourceReplicas
			settings.DataSourceSearchReplicas = dataSourceSearchReplicas
			store := sqlstore.New(*settings, nil)
			defer func() {
				store.Close()
				CleanupSqlSettings(settings)
			}()

			assert.Len(t, store.GetAllConns(), testCase.ExpectedNumConnections)
		})
	}
}

func TestIsDuplicate(t *testing.T) {
	testErrors := map[error]bool{
		&pq.Error{Code: "42P06"}:                                   false,
		&pq.Error{Code: sqlstore.PGDupTableErrorCode}:              true,
		&mysql.MySQLError{Number: uint16(1000)}:                    false,
		&mysql.MySQLError{Number: sqlstore.MySQLDupTableErrorCode}: true,
		errors.New("Random error"):                                 false,
	}

	for err, expected := range testErrors {
		t.Run(fmt.Sprintf("Should return %t for %s", expected, err.Error()), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, expected, sqlstore.IsDuplicate(err))
		})
	}
}

func TestVersionString(t *testing.T) {
	versions := []struct {
		input  int
		output string
	}{
		{
			input:  100000,
			output: "10.0",
		},
		{
			input:  90603,
			output: "9.603",
		},
		{
			input:  120005,
			output: "12.5",
		},
	}

	for _, v := range versions {
		out := sqlstore.VersionString(v.input)
		assert.Equal(t, v.output, out)
	}
}

func makeSqlSettings(driver string) *model.SqlSettings {
	switch driver {
	case model.DATABASE_DRIVER_POSTGRES:
		return MakeSqlSettings(driver, false)
	case model.DATABASE_DRIVER_MYSQL:
		return MakeSqlSettings(driver, false)
	}

	return nil
}

func StoreTestSuiteWithSqlSupplier(t *testing.T, testSuite StoreTestBaseSuite, executeFunc func(t *testing.T, testSuite StoreTestBaseSuite)) {
	for _, st := range StoreTypes {
		st := st
		t.Run(st.Name, func(t *testing.T) {
			if testing.Short() {
				t.SkipNow()
			}
			testSuite.SetStore(st.Store)
			testSuite.SetSqlStore(st.SqlStore)
			// suite.Run(t, testSuite)
			executeFunc(t, testSuite)
		})
	}
}
