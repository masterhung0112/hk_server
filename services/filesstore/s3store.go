package filesstore

import (
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
  return nil
}