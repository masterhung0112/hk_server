package sqlstore

import (
	"os"
	"sync"
	"github.com/masterhung0112/go_server/store/storetest"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"testing"
)

type storeType struct {
  Name string
  SqlSettings *model.SqlSettings
  SqlSupplier *SqlSupplier
  Store store.Store
}

var storeTypes []*storeType
var tearDownStoresOnce sync.Once

func initStores() {
  if testing.Short() {
		return
  }

  // In CI, we already run the entire test suite for both mysql and postgres in parallel.
	// So we just run the tests for the current database set.
	if os.Getenv("IS_CI") == "true" {
    panic("Not implement IS_CI yet")
  } else {
    storeTypes = append(storeTypes, &storeType{
			Name:        "MySQL",
			SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_MYSQL),
		})
		storeTypes = append(storeTypes, &storeType{
			Name:        "PostgreSQL",
			SqlSettings: storetest.MakeSqlSettings(model.DATABASE_DRIVER_POSTGRES),
		})
  }

  defer func() {
		if err := recover(); err != nil {
			tearDownStores()
			panic(err)
		}
  }()

  var wg sync.WaitGroup
  for _, st := range storeTypes {
    st := st
    wg.Add(1)
    go func() {
      defer wg.Done()
      st.SqlSupplier = NewSqlSupplier(*st.SqlSettings)
      st.Store = st.SqlSupplier
      st.Store.DropAllTables()
      st.Store.MarkSystemRanUnitTests()
    }()
  }
  wg.Wait()
}

func tearDownStores() {
  if testing.Short() {
		return
  }
  tearDownStoresOnce.Do(func() {
    var wg sync.WaitGroup
    wg.Add(len(storeTypes))
    for _, st := range storeTypes {
      st := st
      go func() {
        if st.Store != nil {
          st.Store.Close()
        }
        if st.SqlSettings != nil {
					storetest.CleanupSqlSettings(st.SqlSettings)
				}
        wg.Done()
      }()
      wg.Wait()
    }
  })
}

func StoreTestWithSqlSupplier(t *testing.T, f func(*testing.T, store.Store, storetest.SqlSupplier)) {
  for _, st := range storeTypes {
    st := st
    t.Run(st.Name, func(t *testing.T) {
      if testing.Short() {
        t.SkipNow()
      }
      f(t, st.Store, st.SqlSupplier)
    })
  }
}