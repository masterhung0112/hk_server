package filesstore

import (
	"github.com/masterhung0112/go_server/model"
	"io"
	"net/http"
)

type ReadCloseSeeker interface {
	io.ReadCloser
	io.Seeker
}

type FileBackend interface {
	TestConnection() *model.AppError
	WriteFile(fr io.Reader, path string) (int64, *model.AppError)
	RemoveFile(path string) *model.AppError
	Reader(path string) (ReadCloseSeeker, *model.AppError)
	ReadFile(path string) ([]byte, *model.AppError)
}

func NewFileBackend(settings *model.FileSettings, enableComplianceFeatures bool) (FileBackend, *model.AppError) {
	switch *settings.DriverName {
	case model.IMAGE_DRIVER_S3:
		return &S3FileBackend{
			endpoint:   *settings.S3Endpoint,
			accessKey:  *settings.S3AccessKeyId,
			secretKey:  *settings.S3SecretAccessKey,
			secure:     settings.S3SSL == nil || *settings.S3SSL,
			signV2:     settings.S3SignV2 != nil && *settings.S3SignV2,
			region:     *settings.S3Region,
			trace:      settings.S3Trace != nil && *settings.S3Trace,
			bucket:     *settings.S3Bucket,
			pathPrefix: *settings.S3PathPrefix,
			encrypt:    settings.S3SSE != nil && *settings.S3SSE && enableComplianceFeatures,
		}, nil
	case model.IMAGE_DRIVER_LOCAL:
		return &LocalFileBackend{
			directory: *settings.Directory,
		}, nil
	}
	return nil, model.NewAppError("NewFileBackend", "api.file.no_driver.app_error", nil, "", http.StatusInternalServerError)
}
