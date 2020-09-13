package api

import (
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/testlib"
	"testing"
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
