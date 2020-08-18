package filesstore

import (
	"net/http"
	"github.com/masterhung0112/go_server/model"
)

type FileBackend interface {
  TestConnection() *model.AppError
}

func NewFileBackend(settings *model.FileSettings, enableCompilanceFeatures bool) (FileBackend, *model.AppError) {
	switch *settings.DriverName {
	case model.IMAGE_DRIVER_S3:
		return &S3FileBackend{
      endpoint: *settings.S3Endpoint,
      accessKey: *settings.S3AccessKeyId,
      secretKey: *settings.S3SecretAccessKey,
      secure: settings.S3SSL == nil || *settings.S3SSL,
      signV2: settings.S3SignV2 != nil && *settings.S3SignV2,
      region: *settings.S3Region,
      trace: *settings.S3Trace,
    }, nil
	case model.IMAGE_DRIVER_LOCAL:
		return &LocalFileBackend{
      directory: *settings.Directory,
    }, nil
  }
  return nil, model.NewAppError("NewFileBackend", "api.file.no_driver.app_error", nil, "", http.StatusInternalServerError)
}
