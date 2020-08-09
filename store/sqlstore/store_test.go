package sqlstore

import (
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

func StoreTestWithSqlSupplier(t *testing.T, f func(*testing.T, store.Store, storetest.SqlSupplier)) {
  for _, st := range storeTypes {
    st := st
    t.Run(st.name, func(t *testing.T) {
      if testing.Short() {
        t.SkipNow()
      }
      f(t, st.store, st.SqlSupplier)
    })
  }
}