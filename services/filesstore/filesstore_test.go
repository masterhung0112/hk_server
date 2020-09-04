package filesstore

import (
	"bytes"
	"fmt"
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

func TestS3FileBackendTestSuite(t *testing.T) {
	runBackendTest(t, false)
}

func runBackendTest(t *testing.T, encrypt bool) {
	s3Host := os.Getenv("CI_MINIO_HOST")
	if s3Host == "" {
		s3Host = "localhost"
	}

	s3Port := os.Getenv("CI_MINIO_PORT")
	if s3Port == "" {
		s3Port = "9001"
	}

	s3Endpoint := fmt.Sprintf("%s:%s", s3Host, s3Port)

	suite.Run(t, &FileBackendTestSuite{
		settings: model.FileSettings{
			DriverName:        model.NewString(model.IMAGE_DRIVER_S3),
			S3AccessKeyId:     model.NewString(model.MINIO_ACCESS_KEY),
			S3SecretAccessKey: model.NewString(model.MINIO_SECRET_KEY),
			S3Bucket:          model.NewString(model.MINIO_BUCKET),
			S3Region:          model.NewString(""),
			S3Endpoint:        model.NewString(s3Endpoint),
			S3PathPrefix:      model.NewString(""),
			S3SSL:             model.NewBool(false),
			S3SSE:             model.NewBool(encrypt),
		},
	})
}

func (s *FileBackendTestSuite) TestReadWriteFile() {
  b := []byte("test")
  path := "tests/" + model.NewId()

  written, err := s.backend.WriteFile(bytes.NewReader(b), path)
	s.Nil(err)
	s.EqualValues(len(b), written, "expected given number of bytes to have been written")
	defer s.backend.RemoveFile(path)

	read, err := s.backend.ReadFile(path)
	s.Nil(err)

	readString := string(read)
	s.EqualValues(readString, "test")
}