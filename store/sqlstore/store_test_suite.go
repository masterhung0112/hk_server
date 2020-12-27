package sqlstore

import (
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/store"
	"github.com/stretchr/testify/suite"
	"testing"
)

type StoreTestBaseSuite interface {
	suite.TestingSuite

	InitInitializeStore()

	SetStore(store store.Store)
	Store() store.Store

	SetSqlStore(sqlStore *SqlStore)
	SqlStore() *SqlStore
}

type StoreTestSuite struct {
	// suite.Suite

	store    store.Store
	sqlStore *SqlStore
	// sqlSupplier storetest.SqlSupplier
}

/***
 * StoreTestSuite implements interface StoreTestBaseSuite
 ***/
func (s *StoreTestSuite) InitInitializeStore() {
	if len(StoreTypes) >= 1 && (s.Store() == nil || s.SqlStore() == nil) {
		s.SetStore(StoreTypes[0].Store)
		s.SetSqlStore(StoreTypes[0].SqlStore)
	}
}

func (s *StoreTestSuite) SetStore(store store.Store) {
	s.store = store
}

func (s *StoreTestSuite) Store() store.Store {
	return s.store
}

func (s *StoreTestSuite) SetSqlStore(sqlStore *SqlStore) {
	s.sqlStore = sqlStore
}

func (s *StoreTestSuite) SqlStore() *SqlStore {
	return s.sqlStore
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

func StoreTestMysqlTestSuite(t *testing.T, testSuite *suite.Suite) {
	// Setup a global logger to catch tests logging outside of app context
	// The global logger will be stomped by apps initializing but that's fine for testing. Ideally this won't happen.
	mlog.InitGlobalLogger(mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	}))

	// dir, err := ioutil.TempDir("", "")
	// require.NoError(t, err)
	// defer os.RemoveAll(dir)

	suite.Run(t, testSuite)

	// suite.Run(t, &FileBackendTestSuite{
	// 	settings: model.FileSettings{
	// 		DriverName: model.NewString(model.IMAGE_DRIVER_LOCAL),
	// 		Directory:  &dir,
	// 	},
	// })
}
