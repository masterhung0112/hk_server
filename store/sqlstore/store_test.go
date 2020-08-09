package sqlstore

import (
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