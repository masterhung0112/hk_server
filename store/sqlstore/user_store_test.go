package sqlstore

import (
	"github.com/masterhung0112/go_server/store/storetest"
	"testing"
)

func TestUserStore(t *testing.T) {
  StoreTestWithSqlSupplier(t, storetest.TestUserStore)
}