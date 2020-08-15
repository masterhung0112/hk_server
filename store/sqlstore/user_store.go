package sqlstore

import (
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
)

type SqlUserStore struct {
	SqlStore

	// usersQuery is a starting point for all queries that return one or more Users.
	usersQuery sq.SelectBuilder
}

func newSqlUserStore(sqlStore SqlStore) store.UserStore {
	us := &SqlUserStore{
		SqlStore: sqlStore,
	}

	// note: we are providing field names explicitly here to maintain order of columns (needed when using raw queries)
	us.usersQuery = us.getQueryBuilder().
    // Select("u.Id", "u.CreateAt", "u.UpdateAt", "u.DeleteAt", "u.Username", "u.Password", "u.AuthData", "u.AuthService", "u.Email", "u.EmailVerified", "u.Nickname", "u.FirstName", "u.LastName", "u.Position", "u.Roles", "u.AllowMarketing", "u.Props", "u.NotifyProps", "u.LastPasswordUpdate", "u.LastPictureUpdate", "u.FailedAttempts", "u.Locale", "u.Timezone", "u.MfaActive", "u.MfaSecret",
    Select("u.Id", "u.CreateAt", "u.UpdateAt", "u.DeleteAt", "u.Username", "u.Password","u.Email", "u.EmailVerified","u.FirstName", "u.LastName", "u.Roles",
      // "b.UserId IS NOT NULL AS IsBot", "COALESCE(b.Description, '') AS BotDescription", "COALESCE(b.LastIconUpdate, 0) AS BotLastIconUpdate"
    ).
		From("Users u")
		// LeftJoin("Bots b ON ( b.UserId = u.Id )")

  for _, db := range sqlStore.GetAllConns() {
    // Create table users
    table := db.AddTableWithName(model.User{}, "Users").SetKeys(false, "Id")

    // Set constraints for all columns
    table.ColMap("Id").SetMaxSize(26)
    table.ColMap("Username").SetMaxSize(64).SetUnique(true)
    table.ColMap("Password").SetMaxSize(128)
    // table.ColMap("AuthData").SetMaxSize(128).SetUnique(true)
		// table.ColMap("AuthService").SetMaxSize(32)
    table.ColMap("Email").SetMaxSize(128).SetUnique(true)
    // table.ColMap("Nickname").SetMaxSize(64)
    table.ColMap("FirstName").SetMaxSize(64)
    table.ColMap("LastName").SetMaxSize(64)
    table.ColMap("Roles").SetMaxSize(256)
    // table.ColMap("Props").SetMaxSize(4000)
		// table.ColMap("NotifyProps").SetMaxSize(2000)
		// table.ColMap("Locale").SetMaxSize(5)
		// table.ColMap("MfaSecret").SetMaxSize(128)
		// table.ColMap("Position").SetMaxSize(128)
		// table.ColMap("Timezone").SetMaxSize(256)
  }

	return us
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
