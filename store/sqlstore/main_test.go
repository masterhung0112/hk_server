package sqlstore_test

import (
	"github.com/masterhung0112/go_server/store/sqlstore"
	"testing"
)


func TestMain(m *testing.M) {
  mainHelper = testlib.NewMainHelperWithOptions(nil)
  sqlstore.InitTest()

  mainHelper.Main(m)
  sqlstore.TearDownTest()
}