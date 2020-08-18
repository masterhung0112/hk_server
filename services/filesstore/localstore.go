package filesstore

import (
	"github.com/masterhung0112/go_server/model"
)

type LocalFileBackend struct {
	directory string
}

func (b *LocalFileBackend) TestConnection() *model.AppError {
  return nil
}