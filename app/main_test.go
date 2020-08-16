package app

import (
	"testing"
	"github.com/masterhung0112/go_server/testlib"
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