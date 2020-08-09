package sqlstore

import (
	"net/http"
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

func (us SqlUserStore) GetAll() ([]*model.User, *model.AppError) {
  query := us.usersQuery.OrderBy("Username ASC")

  queryString, args, err := query.ToSql()
  if err != nil {
    return nil, model.NewAppError("SqlUserStore.GetAll", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
  }

  var data []*model.User
  if _, err := us.GetReplica().Select(&data, queryString, args...); err != nil {
    return nil, model.NewAppError("SqlUserStore.GetAll", "store.sql_user.get.app_error", nil, err.Error(), http.StatusInternalServerError)
  }

  return data, nil
}

func (us SqlUserStore) PermanentDelete(userId string) *model.AppError {
	if _, err := us.GetMaster().Exec("DELETE FROM Users WHERE Id = :UserId", map[string]interface{}{"UserId": userId}); err != nil {
		return model.NewAppError("SqlUserStore.PermanentDelete", "store.sql_user.permanent_delete.app_error", nil, "userId="+userId+", "+err.Error(), http.StatusInternalServerError)
	}
	return nil
}