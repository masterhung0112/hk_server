package config_test

import (
	"github.com/masterhung0112/go_server/config"
	"github.com/stretchr/testify/require"
	"github.com/masterhung0112/go_server/model"
	"testing"

)

func setupConfigFile(t *testing.T, cfg *model.Config) (string, func()) {
  return "", nil
}

// getActualFileConfig returns the configuration present in the given file without relying on a config store.
func getActualFileConfig(t *testing.T, path string) *model.Config {
  return nil
}

// assertFileEqualsConfig verifies the on disk contents of the given path equal the given config.
func assertFileEqualsConfig(t *testing.T, expectedCfg *model.Config, path string) {
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