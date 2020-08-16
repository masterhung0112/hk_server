package sqlstore

import (
	"testing"

	"github.com/masterhung0112/go_server/store/storetest"
)

func TestUserStore(t *testing.T) {
	StoreTestWithSqlSupplier(t, storetest.TestUserStore)
}
