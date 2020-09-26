package app

import (
	"github.com/masterhung0112/hk_server/testlib"
	"testing"
)

var mainHelper *testlib.MainHelper

func TestMain(m *testing.M) {
	var options = testlib.HelperOptions{
		EnableStore:     true,
		EnableResources: true,
	}

	mainHelper = testlib.NewMainHelperWithOptions(&options)
	defer mainHelper.Close()

	mainHelper.Main(m)
}
