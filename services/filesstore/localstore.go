package filesstore

import (
	"net/http"
	"github.com/masterhung0112/go_server/mlog"
	"os"
	"io"
	"path/filepath"
	"bytes"
	"github.com/masterhung0112/go_server/model"
)

const (
	TEST_FILE_PATH = "/testfile"
)

type LocalFileBackend struct {
	directory string
}

func (b *LocalFileBackend) TestConnection() *model.AppError {
  f := bytes.NewReader([]byte("testingwrite"))
  if _, err := writeFileLocally(f, filepath.Join(b.directory, TEST_FILE_PATH)); err != nil{
		return model.NewAppError("TestFileConnection", "api.file.test_connection.local.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
  }
  os.Remove(filepath.Join(b.directory, TEST_FILE_PATH))
	mlog.Debug("Able to write files to local storage.")
  return nil
}

func writeFileLocally(fr io.Reader, path string) (int64, *model.AppError) {
  if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
    directory, _ := filepath.Abs(filepath.Dir(path))
		return 0, model.NewAppError("WriteFile", "api.file.write_file_locally.create_dir.app_error", nil, "directory="+directory+", err="+err.Error(), http.StatusInternalServerError)
  }
  fw, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return 0, model.NewAppError("WriteFile", "api.file.write_file_locally.writing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer fw.Close()
	written, err := io.Copy(fw, fr)
	if err != nil {
		return written, model.NewAppError("WriteFile", "api.file.write_file_locally.writing.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return written, nil
}