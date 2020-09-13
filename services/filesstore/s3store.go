package filesstore

import (
	"bytes"
	"context"
	"github.com/masterhung0112/go_server/mlog"
	"github.com/minio/minio-go/v7/pkg/encrypt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/masterhung0112/go_server/model"
	s3 "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3FileBackend struct {
	endpoint   string
	accessKey  string
	secretKey  string
	secure     bool
	signV2     bool
	region     string
	trace      bool
	bucket     string
	pathPrefix string
	encrypt    bool
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
			Region:        b.region,
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

func (b *S3FileBackend) WriteFile(fr io.Reader, path string) (int64, *model.AppError) {
	s3Client, err := b.s3New()
	if err != nil {
		return 0, model.NewAppError("WriteFile", "api.file.write_file.s3.connection_fail", nil, err.Error(), http.StatusInternalServerError)
	}

	var contentType string
	path = filepath.Join(b.pathPrefix, path)
	// On Windows, filepath.join return "\", we must replace "\" to "/"
	path = filepath.ToSlash(path)
	if ext := filepath.Ext(path); model.IsFileExtImage(ext) {
		contentType = model.GetImageMimeType(ext)
	} else {
		contentType = "binary/octet-stream"
	}

	options := s3PutOptions(b.encrypt, contentType)
	var buf bytes.Buffer
	_, err = buf.ReadFrom(fr)
	if err != nil {
		return 0, model.NewAppError("WriteFile", "api.file.write_file.s3.read_buf_fail", nil, err.Error(), http.StatusInternalServerError)
	}
	info, err := s3Client.PutObject(context.Background(), b.bucket, path, &buf, int64(buf.Len()), options)
	if err != nil {
		return info.Size, model.NewAppError("WriteFile", "api.file.write_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return info.Size, nil
}

func s3PutOptions(encrypted bool, contentType string) s3.PutObjectOptions {
	options := s3.PutObjectOptions{}
	if encrypted {
		options.ServerSideEncryption = encrypt.NewSSE()
	}
	options.ContentType = contentType

	return options
}

func (b *S3FileBackend) RemoveFile(path string) *model.AppError {
	s3Client, err := b.s3New()
	if err != nil {
		return model.NewAppError("RemoveFile", "utils.file.remove_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	// On Windows, filepath.join return "\", we must replace "\" to "/"
	path = filepath.ToSlash(path)
	if err := s3Client.RemoveObject(context.Background(), b.bucket, path, s3.RemoveObjectOptions{
		GovernanceBypass: true,
		VersionID:        "",
	}); err != nil {
		return model.NewAppError("RemoveFile", "utils.file.remove_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (b *S3FileBackend) ReadFile(path string) ([]byte, *model.AppError) {
	s3Client, err := b.s3New()
	if err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	// On Windows, filepath.join return "\", we must replace "\" to "/"
	path = filepath.ToSlash(path)
	minioObject, err := s3Client.GetObject(context.Background(), b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defer minioObject.Close()
	if f, err := ioutil.ReadAll(minioObject); err != nil {
		return nil, model.NewAppError("ReadFile", "api.file.read_file.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	} else {
		return f, nil
	}
}

func (b *S3FileBackend) Reader(path string) (ReadCloseSeeker, *model.AppError) {
	s3Client, err := b.s3New()
	if err != nil {
		return nil, model.NewAppError("Reader", "api.file.reader.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	path = filepath.Join(b.pathPrefix, path)
	// On Windows, filepath.join return "\", we must replace "\" to "/"
	path = filepath.ToSlash(path)
	minioObject, err := s3Client.GetObject(context.Background(), b.bucket, path, s3.GetObjectOptions{})
	if err != nil {
		return nil, model.NewAppError("Reader", "api.file.reader.s3.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return minioObject, nil
}
