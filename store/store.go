package store

import (
	"github.com/masterhung0112/go_server/model"

)

type Store interface {
  User() UserStore
  Close()
}

type UserStore interface {
  Save(user *model.User) (*model.User, *model.AppError)
  GetAll() ([]*model.User, *model.AppError)
  PermanentDelete(userId string) *model.AppError
}