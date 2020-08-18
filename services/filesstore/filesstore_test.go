package filesstore

import (
	"testing"
	"os"
	"io/ioutil"
	"github.com/masterhung0112/go_server/mlog"
	"github.com/stretchr/testify/require"
	"github.com/masterhung0112/go_server/utils"
	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/suite"
)

type FileBackendTestSuite struct {
	suite.Suite

	settings model.FileSettings
	backend  FileBackend
}

func (s *FileBackendTestSuite) SetupTest() {
	utils.TranslationsPreInit()

	backend, err := NewFileBackend(&s.settings, true)
	require.Nil(s.T(), err)
	s.backend = backend
}

func (s *FileBackendTestSuite) TestConnection() {
	s.Nil(s.backend.TestConnection())
}

func TestLocalFileBackendTestSuite(t *testing.T) {
  // Setup a global logger to catch tests logging outside of app context
	// The global logger will be stomped by apps initializing but that's fine for testing. Ideally this won't happen.
	mlog.InitGlobalLogger(mlog.NewLogger(&mlog.LoggerConfiguration{
		EnableConsole: true,
		ConsoleJson:   true,
		ConsoleLevel:  "error",
		EnableFile:    false,
	}))

	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	suite.Run(t, &FileBackendTestSuite{
		settings: model.FileSettings{
			DriverName: model.NewString(model.IMAGE_DRIVER_LOCAL),
			Directory:  &dir,
		},
	})
}