package main

import (
	"flag"
	"github.com/masterhung0112/hk_server/api"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/testlib"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Command tests are run by re-invoking the test binary in question, so avoid creating
	// another container when we detect same.
	flag.Parse()
	if filter := flag.Lookup("test.run").Value.String(); filter == "ExecCommand" {
		status := m.Run()
		os.Exit(status)
		return
	}

	var options = testlib.HelperOptions{
		EnableStore:     true,
		EnableResources: true,
	}

	mlog.DisableZap()

	mainHelper = testlib.NewMainHelperWithOptions(&options)
	defer mainHelper.Close()
	api.SetMainHelper(mainHelper)

	mainHelper.Main(m)
}
