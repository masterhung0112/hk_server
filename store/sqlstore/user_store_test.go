package sqlstore

import (
	"testing"
)

func TestUserStore(t *testing.T) {
  StoreTestWithSqlSupplier(t, storetest.TestUserStore)
}