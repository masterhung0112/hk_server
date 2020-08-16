package store

import (
	"github.com/masterhung0112/go_server/model"
)

type Store interface {
	User() UserStore
	Close()
	DropAllTables()
	MarkSystemRanUnitTests()
}

type UserStore interface {
	Save(user *model.User) (*model.User, *model.AppError)
	Get(id string) (*model.User, *model.AppError)
	GetAll() ([]*model.User, *model.AppError)
	Count(options model.UserCountOptions) (int64, *model.AppError)
	PermanentDelete(userId string) *model.AppError
}
