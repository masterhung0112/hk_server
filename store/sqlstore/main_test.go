package sqlstore_test

import (
	"github.com/masterhung0112/go_server/testlib"
	"github.com/masterhung0112/go_server/store/sqlstore"
	"testing"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
  mainHelper = testlib.NewMainHelperWithOptions(nil)
  defer mainHelper.Close()
  sqlstore.InitTest()

  mainHelper.Main(m)
  sqlstore.TearDownTest()
}