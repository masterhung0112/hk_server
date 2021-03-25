package sqlstore_test

import (
	"testing"

	"github.com/masterhung0112/hk_server/v5/shared/mlog"
	"github.com/masterhung0112/hk_server/v5/store/sqlstore"
	"github.com/masterhung0112/hk_server/v5/testlib"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	mlog.DisableZap()
	mainHelper = testlib.NewMainHelperWithOptions(nil)
	defer mainHelper.Close()

	sqlstore.InitTest()

	mainHelper.Main(m)
	sqlstore.TearDownTest()
}
