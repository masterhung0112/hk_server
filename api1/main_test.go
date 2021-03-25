package api1

import (
	"testing"

	"github.com/masterhung0112/hk_server/shared/mlog"
	"github.com/masterhung0112/hk_server/testlib"
)

func TestMain(m *testing.M) {
	var options = testlib.HelperOptions{
		EnableStore:     true,
		EnableResources: true,
	}

	mlog.DisableZap()

	mainHelper = testlib.NewMainHelperWithOptions(&options)
	defer mainHelper.Close()

	mainHelper.Main(m)
}
