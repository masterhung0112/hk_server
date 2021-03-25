package storetest_test

import (
	"testing"

	"github.com/masterhung0112/hk_server/shared/mlog"
	"github.com/masterhung0112/hk_server/store/storetest"
	"github.com/masterhung0112/hk_server/testlib"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	mlog.DisableZap()
	mainHelper = testlib.NewMainHelperWithOptions(nil)
	defer mainHelper.Close()

	storetest.InitTest()

	mainHelper.Main(m)
	storetest.TearDownTest()
}
