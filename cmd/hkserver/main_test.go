package main

import (
	"flag"
	"os"
	"testing"

	api4 "github.com/masterhung0112/hk_server/v5/api4"
	"github.com/masterhung0112/hk_server/v5/shared/mlog"
	"github.com/masterhung0112/hk_server/v5/testlib"
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
	api4.SetMainHelper(mainHelper)

	mainHelper.Main(m)
}
