package sqlstore

import (
	"database/sql"
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
		Select("u.Id", "u.CreateAt", "u.UpdateAt", "u.DeleteAt", "u.Username", "u.Password", "u.Email", "u.EmailVerified", "u.FirstName", "u.LastName", "u.Roles"). // "b.UserId IS NOT NULL AS IsBot", "COALESCE(b.Description, '') AS BotDescription", "COALESCE(b.LastIconUpdate, 0) AS BotLastIconUpdate"
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
	if len(user.Id) > 0 {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.existing.app_error", nil, "user_id="+user.Id, http.StatusBadRequest)
	}

	// Fill up, Transform data before save
	user.PreSave()
	if err := user.IsValid(); err != nil {
		return nil, err
	}

	// Save user to master database
	if err := us.GetMaster().Insert(user); err != nil {
		if IsUniqueConstraintError(err, []string{"Email", "users_email_key", "idx_users_email_unique"}) {
			return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.email_exists.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusBadRequest)
		}
		if IsUniqueConstraintError(err, []string{"Username", "users_username_key", "idx_users_username_unique"}) {
			return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.username_exists.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusInternalServerError)
	}

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

func (us SqlUserStore) Get(id string) (*model.User, *model.AppError) {
	failure := func(err error, id string, statusCode int) *model.AppError {
		details := "user_id=" + id + ", " + err.Error()
		return model.NewAppError("SqlUserStore.Get", id, nil, details, statusCode)
	}

	query := us.usersQuery.Where("Id = ?", id)
	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}
	row := us.GetReplica().Db.QueryRow(queryString, args...)

	var user model.User
	// var props, notifyProps, timezone []byte
	err = row.Scan(&user.Id, &user.CreateAt, &user.UpdateAt, &user.DeleteAt, &user.Username,
		// &user.Password, &user.AuthData, &user.AuthService, &user.Email, &user.EmailVerified,
		// &user.Nickname, &user.FirstName, &user.LastName, &user.Position, &user.Roles,
		// &user.AllowMarketing, &props, &notifyProps, &user.LastPasswordUpdate, &user.LastPictureUpdate,
		// &user.FailedAttempts, &user.Locale, &timezone, &user.MfaActive, &user.MfaSecret,
		&user.Password, &user.Email, &user.EmailVerified,
		&user.FirstName, &user.LastName, &user.Roles)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, failure(err, store.MISSING_ACCOUNT_ERROR, http.StatusNotFound)
		}
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}
	//TODO: open
	// if err = json.Unmarshal(props, &user.Props); err != nil {
	// 	return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	// }
	// if err = json.Unmarshal(notifyProps, &user.NotifyProps); err != nil {
	// 	return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	// }
	// if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
	// 	return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	// }

	return &user, nil
}

func (us SqlUserStore) Count(options model.UserCountOptions) (int64, *model.AppError) {
  isPostgreSQL := us.DriverName() == model.DATABASE_DRIVER_POSTGRES
	query := us.getQueryBuilder().Select("COUNT(DISTINCT u.Id)").From("Users AS u")

	if !options.IncludeDeleted {
		query = query.Where("u.DeleteAt = 0")
  }

  if options.IncludeBotAccounts {
		if options.ExcludeRegularUsers {
			query = query.Join("Bots ON u.Id = Bots.UserId")
		}
	} else {
    //TODO: Open
		// query = query.LeftJoin("Bots ON u.Id = Bots.UserId").Where("Bots.UserId IS NULL")
		if options.ExcludeRegularUsers {
			// Currently this doesn't make sense because it will always return 0
			return int64(0), model.NewAppError("SqlUserStore.Count", "store.sql_user.count.app_error", nil, "", http.StatusInternalServerError)
		}
  }

  if options.TeamId != "" {
		query = query.LeftJoin("TeamMembers AS tm ON u.Id = tm.UserId").Where("tm.TeamId = ? AND tm.DeleteAt = 0", options.TeamId)
	} else if options.ChannelId != "" {
		query = query.LeftJoin("ChannelMembers AS cm ON u.Id = cm.UserId").Where("cm.ChannelId = ?", options.ChannelId)
  }
  // query = applyViewRestrictionsFilter(query, options.ViewRestrictions, false)
  // query = applyMultiRoleFilters(query, options.Roles, options.TeamRoles, options.ChannelRoles)

  if isPostgreSQL {
		query = query.PlaceholderFormat(sq.Dollar)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return int64(0), model.NewAppError("SqlUserStore.Get", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := us.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return int64(0), model.NewAppError("SqlUserStore.Count", "store.sql_user.get_total_users_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return count, nil

}

func (us SqlUserStore) InferSystemInstallDate() (int64, *model.AppError) {
	createAt, err := us.GetReplica().SelectInt("SELECT CreateAt FROM Users WHERE CreateAt IS NOT NULL ORDER BY CreateAt ASC LIMIT 1")
	if err != nil {
		return 0, model.NewAppError("SqlUserStore.GetSystemInstallDate", "store.sql_user.get_system_install_date.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return createAt, nil
}