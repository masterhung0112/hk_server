package sqlstore_test

import (
	"github.com/masterhung0112/hk_server/store/sqlstore"
	"github.com/masterhung0112/hk_server/testlib"
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
