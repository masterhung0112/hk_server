package storetest_test

import (
	"testing"

	"github.com/masterhung0112/hk_server/v5/shared/mlog"
	"github.com/masterhung0112/hk_server/v5/store/storetest"
	"github.com/masterhung0112/hk_server/v5/testlib"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	mlog.DisableZap()
	mainHelper = testlib.NewMainHelperWithOptions(nil)
	defer mainHelper.Close()

	storetest.InitTestEnv()

	mainHelper.Main(m)
	storetest.TearDownTest()
}
