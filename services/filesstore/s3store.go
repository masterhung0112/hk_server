package filesstore

import (
	"context"
	"github.com/masterhung0112/go_server/mlog"
	"net/http"
	"os"

	"github.com/masterhung0112/go_server/model"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3FileBackend struct {
	endpoint  string
	accessKey string
	secretKey string
	secure    bool
	signV2    bool
	region    string
  trace     bool
  bucket     string
	pathPrefix string
}

func (b *S3FileBackend) s3New() (*s3.Client, error) {
	var creds *credentials.Credentials

	if b.accessKey == "" && b.secretKey == "" {
		creds = credentials.NewIAM("")
	} else if b.signV2 {
		creds = credentials.NewStaticV2(b.accessKey, b.secretKey, "")
	} else {
		creds = credentials.NewStaticV4(b.accessKey, b.secretKey, "")
	}

	s3Client, err := s3.New(b.endpoint, &s3.Options{
		Creds:  creds,
		Secure: b.secure,
		Region: b.region,
	})
	if err != nil {
		return nil, err
	}

	if b.trace {
		s3Client.TraceOn(os.Stdout)
	}

	return s3Client, nil
}

func (b *S3FileBackend) TestConnection() *model.AppError {
  s3Client, err := b.s3New()

  if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.connection.app_error", nil, err.Error(), http.StatusInternalServerError)
  }

  exists, err := s3Client.BucketExists(context.Background(), b.bucket)
	if err != nil {
		return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.bucket_exists.app_error", nil, err.Error(), http.StatusInternalServerError)
  }

  if !exists {
		mlog.Warn("Bucket specified does not exist. Attempting to create...")
		err := s3Client.MakeBucket(context.Background(), b.bucket, s3.MakeBucketOptions{
      Region: b.region,
      ObjectLocking: false,
    })
		if err != nil {
			mlog.Error("Unable to create bucket.")
			return model.NewAppError("TestFileConnection", "api.file.test_connection.s3.bucked_create.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	mlog.Debug("Connection to S3 or minio is good. Bucket exists.")
  return nil
}