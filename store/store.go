package store

import (
	"github.com/masterhung0112/go_server/model"
)

type Store interface {
  User() UserStore
  System() SystemStore
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

type SystemStore interface {
	Save(system *model.System) error
	SaveOrUpdate(system *model.System) error
	Update(system *model.System) error
	Get() (model.StringMap, error)
	GetByName(name string) (*model.System, error)
	PermanentDeleteByName(name string) (*model.System, error)
	InsertIfExists(system *model.System) (*model.System, error)
}