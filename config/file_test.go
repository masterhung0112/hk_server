package config_test

import (
	"fmt"
	"github.com/masterhung0112/go_server/config"
	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

func setupConfigFile(t *testing.T, cfg *model.Config) (string, func()) {
	os.Clearenv()
	t.Helper()

	tempDir, err := ioutil.TempDir("", "setupConfigFile")
	require.NoError(t, err)

	err = os.Chdir(tempDir)
	require.NoError(t, err)

	var name string
	if cfg != nil {
		f, err := ioutil.TempFile(tempDir, "setupConfigFile")
		require.NoError(t, err)

		cfgData, err := config.MarshalConfig(cfg)
		require.NoError(t, err)

		ioutil.WriteFile(f.Name(), cfgData, 0644)

		name = f.Name()
		fmt.Printf("Write to file %s\n", name)
	}

	return name, func() {
		os.RemoveAll(tempDir)
	}
}

// getActualFileConfig returns the configuration present in the given file without relying on a config store.
func getActualFileConfig(t *testing.T, path string) *model.Config {
	t.Helper()

	f, err := os.Open(path)
	require.Nil(t, err)
	defer f.Close()

	actualCfg, _, err := config.UnmarshalConfig(f, false)

	require.Nil(t, err)

	return actualCfg
}

// assertFileEqualsConfig verifies the on disk contents of the given path equal the given config.
func assertFileEqualsConfig(t *testing.T, expectedCfg *model.Config, path string) {
	t.Helper()

	expectedCfg = prepareExpectedConfig(t, expectedCfg)
	actualCfg := getActualFileConfig(t, path)

	assert.Equal(t, expectedCfg, actualCfg)
}

// assertFileNotEqualsConfig verifies the on disk contents of the given path does not equal the given config.
func assertFileNotEqualsConfig(t *testing.T, expectedCfg *model.Config, path string) {

}

func TestFileStoreNew(t *testing.T) {
	t.Run("Absolute path, initialization required", func(t *testing.T) {
		path, tearDown := setupConfigFile(t, testConfig)
		defer tearDown()

		fs, err := config.NewFileStore(path, false)
		require.NoError(t, err)
		defer fs.Close()

		// assert.Equal(t, "http://TestStoreNew", *fs.Get().ServiceSettings.SiteURL)
		assertFileNotEqualsConfig(t, testConfig, path)
	})

	t.Run("Absolute path, already minimally configured", func(t *testing.T) {
		path, tearDown := setupConfigFile(t, minimalConfig)
		defer tearDown()

		fs, err := config.NewFileStore(path, false)
		require.NoError(t, err)
		defer fs.Close()

		// assert.Equal(t, "http://minimal", *fs.Get().ServiceSettings.SiteURL)
		assertFileEqualsConfig(t, minimalConfig, path)
	})

	t.Run("Absolute path, file does not exist", func(t *testing.T) {

	})

	t.Run("Absolute path, path to file does not exist", func(t *testing.T) {

	})

	t.Run("file exists", func(t *testing.T) {

	})

	t.Run("file does not exist", func(t *testing.T) {

	})
}
