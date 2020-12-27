package main

import (
	api1 "github.com/masterhung0112/hk_server/api1"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/testlib"
	"io/ioutil"
	"path/filepath"
	"testing"
)

var mainHelper *testlib.MainHelper

type testHelper struct {
	*api.TestHelper

	config            *model.Config
	tempDir           string
	configFilePath    string
	disableAutoConfig bool
}

// Setup creates an instance of testHelper.
func Setup(t testing.TB) *testHelper {
	dir, err := ioutil.TempDir("", "testHelper")
	if err != nil {
		panic("failed to create temporary directory: " + err.Error())
	}

	api1TestHelper := api.Setup(t)

	testHelper := &testHelper{
		TestHelper:     api1TestHelper,
		tempDir:        dir,
		configFilePath: filepath.Join(dir, "config-helper.json"),
	}

	config := &model.Config{}
	config.SetDefaults()
	testHelper.SetConfig(config)

	return testHelper
}

// SetConfig replaces the configuration passed to a running command.
func (h *testHelper) SetConfig(config *model.Config) {
	if !testing.Short() {
		config.SqlSettings = *mainHelper.GetSQLSettings()
	}

	// Disable strict password requirements for test
	*config.PasswordSettings.MinimumLength = 5
	*config.PasswordSettings.Lowercase = false
	*config.PasswordSettings.Uppercase = false
	*config.PasswordSettings.Symbol = false
	*config.PasswordSettings.Number = false

	h.config = config

	if err := ioutil.WriteFile(h.configFilePath, []byte(config.ToJson()), 0600); err != nil {
		panic("failed to write file " + h.configFilePath + ": " + err.Error())
	}
}
