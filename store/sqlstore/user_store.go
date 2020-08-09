package sqlstore

import (
	"github.com/masterhung0112/go_server/model"
  sq "github.com/Masterminds/squirrel"
)

type SqlUserStore struct {
  SqlStore

  // usersQuery is a starting point for all queries that return one or more Users.
  usersQuery sq.SelectBuilder
}

func (us SqlUserStore) Save(user *model.User) (*model.User, *model.AppError) {

  return user, nil
}
