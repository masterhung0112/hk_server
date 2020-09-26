package sqlstore

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
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
		Select("u.Id", "u.CreateAt", "u.UpdateAt", "u.DeleteAt", "u.Username", "u.Password", "u.AuthData", "u.AuthService", "u.Email", "u.EmailVerified", "u.Nickname", "u.FirstName", "u.LastName", "u.Position", "u.Roles", "u.AllowMarketing", "u.Props", "u.NotifyProps", "u.LastPasswordUpdate", "u.LastPictureUpdate", "u.FailedAttempts", "u.Locale", "u.Timezone", "u.MfaActive", "u.MfaSecret").
		// "b.UserId IS NOT NULL AS IsBot", "COALESCE(b.Description, '') AS BotDescription", "COALESCE(b.LastIconUpdate, 0) AS BotLastIconUpdate").
		From("Users u")
	// LeftJoin("Bots b ON ( b.UserId = u.Id )")

	for _, db := range sqlStore.GetAllConns() {
		// Create table users
		table := db.AddTableWithName(model.User{}, "Users").SetKeys(false, "Id")

		// Set constraints for all columns
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Username").SetMaxSize(64).SetUnique(true)
		table.ColMap("Password").SetMaxSize(128)
		table.ColMap("AuthData").SetMaxSize(128).SetUnique(true)
		table.ColMap("AuthService").SetMaxSize(32)
		table.ColMap("Email").SetMaxSize(128).SetUnique(true)
		table.ColMap("Nickname").SetMaxSize(64)
		table.ColMap("FirstName").SetMaxSize(64)
		table.ColMap("LastName").SetMaxSize(64)
		table.ColMap("Roles").SetMaxSize(256)
		table.ColMap("Props").SetMaxSize(4000)
		table.ColMap("NotifyProps").SetMaxSize(2000)
		table.ColMap("Locale").SetMaxSize(5)
		table.ColMap("MfaSecret").SetMaxSize(128)
		table.ColMap("Position").SetMaxSize(128)
		table.ColMap("Timezone").SetMaxSize(256)
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
	var props, notifyProps, timezone []byte
	err = row.Scan(&user.Id, &user.CreateAt, &user.UpdateAt, &user.DeleteAt, &user.Username,
		&user.Password, &user.AuthData, &user.AuthService, &user.Email, &user.EmailVerified,
		&user.Nickname, &user.FirstName, &user.LastName, &user.Position, &user.Roles,
		&user.AllowMarketing, &props, &notifyProps, &user.LastPasswordUpdate, &user.LastPictureUpdate,
		&user.FailedAttempts, &user.Locale, &timezone, &user.MfaActive, &user.MfaSecret)
	// &user.IsBot, &user.BotDescription, &user.BotLastIconUpdate)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, failure(err, store.MISSING_ACCOUNT_ERROR, http.StatusNotFound)
		}
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}

	if err = json.Unmarshal(props, &user.Props); err != nil {
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}
	if err = json.Unmarshal(notifyProps, &user.NotifyProps); err != nil {
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}
	if err = json.Unmarshal(timezone, &user.Timezone); err != nil {
		return nil, failure(err, "store.sql_user.get.app_error", http.StatusInternalServerError)
	}

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

func (us SqlUserStore) GetByEmail(email string) (*model.User, *model.AppError) {
	query := us.usersQuery.Where("Email = lower(?)", email)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetByEmail", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	user := model.User{}
	if err := us.GetReplica().SelectOne(&user, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetByEmail", store.MISSING_ACCOUNT_ERROR, nil, "email="+email+", "+err.Error(), http.StatusInternalServerError)
	}

	return &user, nil
}

func (us SqlUserStore) GetByUsername(username string) (*model.User, *model.AppError) {
	query := us.usersQuery.Where("u.Username = lower(?)", username)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetByUsername", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var user *model.User
	if err := us.GetReplica().SelectOne(&user, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetByUsername", "store.sql_user.get_by_username.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return user, nil
}

func applyViewRestrictionsFilter(query sq.SelectBuilder, restrictions *model.ViewUsersRestrictions, distinct bool) sq.SelectBuilder {
	if restrictions == nil {
		return query
	}

	// If you have no access to teams or channels, return and empty result.
	if restrictions.Teams != nil && len(restrictions.Teams) == 0 && restrictions.Channels != nil && len(restrictions.Channels) == 0 {
		return query.Where("1 = 0")
	}

	teams := make([]interface{}, len(restrictions.Teams))
	for i, v := range restrictions.Teams {
		teams[i] = v
	}
	channels := make([]interface{}, len(restrictions.Channels))
	for i, v := range restrictions.Channels {
		channels[i] = v
	}
	resultQuery := query
	if restrictions.Teams != nil && len(restrictions.Teams) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("TeamMembers rtm ON ( rtm.UserId = u.Id AND rtm.DeleteAt = 0 AND rtm.TeamId IN (%s))", sq.Placeholders(len(teams))), teams...)
	}
	if restrictions.Channels != nil && len(restrictions.Channels) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("ChannelMembers rcm ON ( rcm.UserId = u.Id AND rcm.ChannelId IN (%s))", sq.Placeholders(len(channels))), channels...)
	}

	if distinct {
		return resultQuery.Distinct()
	}

	return resultQuery
}

func applyRoleFilter(query sq.SelectBuilder, role string, isPostgreSQL bool) sq.SelectBuilder {
	if role == "" {
		return query
	}

	if isPostgreSQL {
		roleParam := fmt.Sprintf("%%%s%%", sanitizeSearchTerm(role, "\\"))
		return query.Where("u.Roles LIKE LOWER(?)", roleParam)
	}

	roleParam := fmt.Sprintf("%%%s%%", sanitizeSearchTerm(role, "*"))

	return query.Where("u.Roles LIKE ? ESCAPE '*'", roleParam)
}

func (us SqlUserStore) GetProfilesWithoutTeam(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	isPostgreSQL := us.DriverName() == model.DATABASE_DRIVER_POSTGRES
	query := us.usersQuery.
		Where(`(
			SELECT
				COUNT(0)
			FROM
				TeamMembers
			WHERE
				TeamMembers.UserId = u.Id
				AND TeamMembers.DeleteAt = 0
		) = 0`).
		OrderBy("u.Username ASC").
		Offset(uint64(options.Page * options.PerPage)).Limit(uint64(options.PerPage))

	query = applyViewRestrictionsFilter(query, options.ViewRestrictions, true)

	query = applyRoleFilter(query, options.Role, isPostgreSQL)

	if options.Inactive {
		query = query.Where("u.DeleteAt != 0")
	} else if options.Active {
		query = query.Where("u.DeleteAt = 0")
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesWithoutTeam", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesWithoutTeam", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func applyChannelGroupConstrainedFilter(query sq.SelectBuilder, channelId string) sq.SelectBuilder {
	if channelId == "" {
		return query
	}

	return query.
		Where(`u.Id IN (
				SELECT
					GroupMembers.UserId
				FROM
					Channels
					JOIN GroupChannels ON GroupChannels.ChannelId = Channels.Id
					JOIN UserGroups ON UserGroups.Id = GroupChannels.GroupId
					JOIN GroupMembers ON GroupMembers.GroupId = UserGroups.Id
				WHERE
					Channels.Id = ?
					AND GroupChannels.DeleteAt = 0
					AND UserGroups.DeleteAt = 0
					AND GroupMembers.DeleteAt = 0
				GROUP BY
					GroupMembers.UserId
			)`, channelId)
}

func (us SqlUserStore) GetProfilesNotInChannel(teamId string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	query := us.usersQuery.
		Join("TeamMembers tm ON ( tm.UserId = u.Id AND tm.DeleteAt = 0 AND tm.TeamId = ? )", teamId).
		LeftJoin("ChannelMembers cm ON ( cm.UserId = u.Id AND cm.ChannelId = ? )", channelId).
		Where("cm.UserId IS NULL").
		OrderBy("u.Username ASC").
		Offset(uint64(offset)).Limit(uint64(limit))

	query = applyViewRestrictionsFilter(query, viewRestrictions, true)

	if groupConstrained {
		query = applyChannelGroupConstrainedFilter(query, channelId)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesNotInChannel", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesNotInChannel", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) GetEtagForProfilesNotInTeam(teamId string) string {
	querystr := `
		SELECT
			CONCAT(MAX(UpdateAt), '.', COUNT(Id)) as etag
		FROM
			Users as u
		LEFT JOIN TeamMembers tm
			ON tm.UserId = u.Id
			AND tm.TeamId = :TeamId
			AND tm.DeleteAt = 0
		WHERE
			tm.UserId IS NULL
	`
	etag, err := us.GetReplica().SelectStr(querystr, map[string]interface{}{"TeamId": teamId})
	if err != nil {
		return fmt.Sprintf("%v.%v", model.CurrentVersion, model.GetMillis())
	}

	return fmt.Sprintf("%v.%v", model.CurrentVersion, etag)
}

func applyTeamGroupConstrainedFilter(query sq.SelectBuilder, teamId string) sq.SelectBuilder {
	if teamId == "" {
		return query
	}

	return query.
		Where(`u.Id IN (
				SELECT
					GroupMembers.UserId
				FROM
					Teams
					JOIN GroupTeams ON GroupTeams.TeamId = Teams.Id
					JOIN UserGroups ON UserGroups.Id = GroupTeams.GroupId
					JOIN GroupMembers ON GroupMembers.GroupId = UserGroups.Id
				WHERE
					Teams.Id = ?
					AND GroupTeams.DeleteAt = 0
					AND UserGroups.DeleteAt = 0
					AND GroupMembers.DeleteAt = 0
				GROUP BY
					GroupMembers.UserId
			)`, teamId)
}

func (us SqlUserStore) GetProfilesNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	var users []*model.User
	query := us.usersQuery.
		LeftJoin("TeamMembers tm ON ( tm.UserId = u.Id AND tm.DeleteAt = 0 AND tm.TeamId = ? )", teamId).
		Where("tm.UserId IS NULL").
		OrderBy("u.Username ASC").
		Offset(uint64(offset)).Limit(uint64(limit))

	query = applyViewRestrictionsFilter(query, viewRestrictions, true)

	if groupConstrained {
		query = applyTeamGroupConstrainedFilter(query, teamId)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesNotInTeam", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesNotInTeam", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}
	return users, nil
}

func (us SqlUserStore) GetEtagForProfiles(teamId string) string {
	updateAt, err := us.GetReplica().SelectInt("SELECT UpdateAt FROM Users, TeamMembers WHERE TeamMembers.TeamId = :TeamId AND Users.Id = TeamMembers.UserId ORDER BY UpdateAt DESC LIMIT 1", map[string]interface{}{"TeamId": teamId})
	if err != nil {
		return fmt.Sprintf("%v.%v", model.CurrentVersion, model.GetMillis())
	}
	return fmt.Sprintf("%v.%v", model.CurrentVersion, updateAt)
}

func applyMultiRoleFilters(query sq.SelectBuilder, roles []string, teamRoles []string, channelRoles []string) sq.SelectBuilder {
	queryString := ""
	if len(roles) > 0 && roles[0] != "" {
		schemeGuest := false
		schemeAdmin := false
		schemeUser := false

		for _, role := range roles {
			switch role {
			case model.SYSTEM_ADMIN_ROLE_ID:
				schemeAdmin = true
			case model.SYSTEM_USER_ROLE_ID:
				schemeUser = true
			case model.SYSTEM_GUEST_ROLE_ID:
				schemeGuest = true
			}
		}

		if schemeAdmin || schemeUser || schemeGuest {
			if schemeAdmin && schemeUser {
				queryString += `(u.Roles LIKE '%system_user%' OR u.Roles LIKE '%system_admin%') `
			} else if schemeAdmin {
				queryString += `(u.Roles LIKE '%system_admin%') `
			} else if schemeUser {
				queryString += `(u.Roles LIKE '%system_user%' AND u.Roles NOT LIKE '%system_admin%') `
			}

			if schemeGuest {
				if queryString != "" {
					queryString += "OR "
				}
				queryString += `(u.Roles LIKE '%system_guest%') `
			}
		}
	}

	if len(channelRoles) > 0 && channelRoles[0] != "" {
		schemeGuest := false
		schemeAdmin := false
		schemeUser := false
		for _, channelRole := range channelRoles {
			switch channelRole {
			case model.CHANNEL_ADMIN_ROLE_ID:
				schemeAdmin = true
			case model.CHANNEL_USER_ROLE_ID:
				schemeUser = true
			case model.CHANNEL_GUEST_ROLE_ID:
				schemeGuest = true
			}
		}

		if schemeAdmin || schemeUser || schemeGuest {
			if queryString != "" {
				queryString += "OR "
			}
			if schemeAdmin && schemeUser {
				queryString += `(cm.SchemeUser = true AND u.Roles = 'system_user')`
			} else if schemeAdmin {
				queryString += `(cm.SchemeAdmin = true AND u.Roles = 'system_user')`
			} else if schemeUser {
				queryString += `(cm.SchemeUser = true AND cm.SchemeAdmin = false AND u.Roles = 'system_user')`
			}

			if schemeGuest {
				if queryString != "" && queryString[len(queryString)-3:] != "OR " {
					queryString += "OR "
				}
				queryString += `(cm.SchemeGuest = true AND u.Roles = 'system_guest')`
			}
		}
	}

	if len(teamRoles) > 0 && teamRoles[0] != "" {
		schemeAdmin := false
		schemeUser := false
		schemeGuest := false
		for _, teamRole := range teamRoles {
			switch teamRole {
			case model.TEAM_ADMIN_ROLE_ID:
				schemeAdmin = true
			case model.TEAM_USER_ROLE_ID:
				schemeUser = true
			case model.TEAM_GUEST_ROLE_ID:
				schemeGuest = true
			}
		}

		if schemeAdmin || schemeUser || schemeGuest {
			if queryString != "" {
				queryString += "OR "
			}
			if schemeAdmin && schemeUser {
				queryString += `(tm.SchemeUser = true AND u.Roles = 'system_user')`
			} else if schemeAdmin {
				queryString += `(tm.SchemeAdmin = true AND u.Roles = 'system_user')`
			} else if schemeUser {
				queryString += `(tm.SchemeUser = true AND tm.SchemeAdmin = false AND u.Roles = 'system_user')`
			}

			if schemeGuest {
				if queryString != "" && queryString[len(queryString)-3:] != "OR " {
					queryString += "OR "
				}
				queryString += `(tm.SchemeGuest = true AND u.Roles = 'system_guest')`
			}
		}
	}

	if queryString != "" {
		query = query.Where("(" + queryString + ")")
	}

	return query
}

func (us SqlUserStore) GetProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	isPostgreSQL := us.DriverName() == model.DATABASE_DRIVER_POSTGRES
	query := us.usersQuery.
		Join("TeamMembers tm ON ( tm.UserId = u.Id AND tm.DeleteAt = 0 )").
		Where("tm.TeamId = ?", options.InTeamId).
		OrderBy("u.Username ASC").
		Offset(uint64(options.Page * options.PerPage)).Limit(uint64(options.PerPage))

	query = applyViewRestrictionsFilter(query, options.ViewRestrictions, true)

	query = applyRoleFilter(query, options.Role, isPostgreSQL)
	query = applyMultiRoleFilters(query, options.Roles, options.TeamRoles, options.ChannelRoles)

	if options.Inactive {
		query = query.Where("u.DeleteAt != 0")
	} else if options.Active {
		query = query.Where("u.DeleteAt = 0")
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfiles", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfiles", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) GetProfilesInChannelByStatus(channelId string, offset int, limit int) ([]*model.User, *model.AppError) {
	query := us.usersQuery.
		Join("ChannelMembers cm ON ( cm.UserId = u.Id )").
		LeftJoin("Status s ON ( s.UserId = u.Id )").
		Where("cm.ChannelId = ?", channelId).
		OrderBy(`
			CASE s.Status
				WHEN 'online' THEN 1
				WHEN 'away' THEN 2
				WHEN 'dnd' THEN 3
				ELSE 4
			END
			`).
		OrderBy("u.Username ASC").
		Offset(uint64(offset)).Limit(uint64(limit))

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesInChannelByStatus", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesInChannelByStatus", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) GetProfilesInChannel(channelId string, offset int, limit int) ([]*model.User, *model.AppError) {
	query := us.usersQuery.
		Join("ChannelMembers cm ON ( cm.UserId = u.Id )").
		Where("cm.ChannelId = ?", channelId).
		OrderBy("u.Username ASC").
		Offset(uint64(offset)).Limit(uint64(limit))

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesInChannel", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfilesInChannel", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) GetAllProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	isPostgreSQL := us.DriverName() == model.DATABASE_DRIVER_POSTGRES
	query := us.usersQuery.
		OrderBy("u.Username ASC").
		Offset(uint64(options.Page * options.PerPage)).Limit(uint64(options.PerPage))

	query = applyViewRestrictionsFilter(query, options.ViewRestrictions, true)

	query = applyRoleFilter(query, options.Role, isPostgreSQL)

	if options.Inactive {
		query = query.Where("u.DeleteAt != 0")
	} else if options.Active {
		query = query.Where("u.DeleteAt = 0")
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetAllProfiles", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	fmt.Println(queryString)
	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetAllProfiles", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) Update(user *model.User, trustedUpdateData bool) (*model.UserUpdate, *model.AppError) {
	user.PreUpdate()

	if err := user.IsValid(); err != nil {
		return nil, err
	}

	oldUserResult, err := us.GetMaster().Get(model.User{}, user.Id)
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.finding.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusInternalServerError)
	}

	if oldUserResult == nil {
		return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.find.app_error", nil, "user_id="+user.Id, http.StatusBadRequest)
	}

	oldUser := oldUserResult.(*model.User)
	user.CreateAt = oldUser.CreateAt
	//TODO: Open this
	// user.AuthData = oldUser.AuthData
	// user.AuthService = oldUser.AuthService
	user.Password = oldUser.Password
	// user.LastPasswordUpdate = oldUser.LastPasswordUpdate
	// user.LastPictureUpdate = oldUser.LastPictureUpdate
	user.EmailVerified = oldUser.EmailVerified
	// user.FailedAttempts = oldUser.FailedAttempts
	// user.MfaSecret = oldUser.MfaSecret
	// user.MfaActive = oldUser.MfaActive

	if !trustedUpdateData {
		user.Roles = oldUser.Roles
		user.DeleteAt = oldUser.DeleteAt
	}

	//TODO: Open this
	// if user.IsOAuthUser() {
	// 	if !trustedUpdateData {
	// 		user.Email = oldUser.Email
	// 	}
	// } else if user.IsLDAPUser() && !trustedUpdateData {
	// 	if user.Username != oldUser.Username || user.Email != oldUser.Email {
	// 		return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.can_not_change_ldap.app_error", nil, "user_id="+user.Id, http.StatusBadRequest)
	// 	}
	// } else
	if user.Email != oldUser.Email {
		user.EmailVerified = false
	}

	//TODO: Open
	// if user.Username != oldUser.Username {
	// 	user.UpdateMentionKeysFromUsername(oldUser.Username)
	// }

	count, err := us.GetMaster().Update(user)
	if err != nil {
		if IsUniqueConstraintError(err, []string{"Email", "users_email_key", "idx_users_email_unique"}) {
			return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.email_taken.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusBadRequest)
		}
		if IsUniqueConstraintError(err, []string{"Username", "users_username_key", "idx_users_username_unique"}) {
			return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.username_taken.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusBadRequest)
		}
		return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.updating.app_error", nil, "user_id="+user.Id+", "+err.Error(), http.StatusInternalServerError)
	}

	if count > 1 {
		return nil, model.NewAppError("SqlUserStore.Update", "store.sql_user.update.app_error", nil, fmt.Sprintf("user_id=%v, count=%v", user.Id, count), http.StatusInternalServerError)
	}

	user.Sanitize(map[string]bool{})
	oldUser.Sanitize(map[string]bool{})
	return &model.UserUpdate{New: user, Old: oldUser}, nil
}

func (us SqlUserStore) UpdateFailedPasswordAttempts(userId string, attempts int) *model.AppError {
	if _, err := us.GetMaster().Exec("UPDATE Users SET FailedAttempts = :FailedAttempts WHERE Id = :UserId", map[string]interface{}{"FailedAttempts": attempts, "UserId": userId}); err != nil {
		return model.NewAppError("SqlUserStore.UpdateFailedPasswordAttempts", "store.sql_user.update_failed_pwd_attempts.app_error", nil, "user_id="+userId, http.StatusInternalServerError)
	}

	return nil
}

func (us SqlUserStore) GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, *model.AppError) {
	query := us.usersQuery
	if allowSignInWithUsername && allowSignInWithEmail {
		query = query.Where("Username = lower(?) OR Email = lower(?)", loginId, loginId)
	} else if allowSignInWithUsername {
		query = query.Where("Username = lower(?)", loginId)
	} else if allowSignInWithEmail {
		query = query.Where("Email = lower(?)", loginId)
	} else {
		return nil, model.NewAppError("SqlUserStore.GetForLogin", "store.sql_user.get_for_login.app_error", nil, "", http.StatusInternalServerError)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetForLogin", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	users := []*model.User{}
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetForLogin", "store.sql_user.get_for_login.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if len(users) == 0 {
		return nil, model.NewAppError("SqlUserStore.GetForLogin", "store.sql_user.get_for_login.app_error", nil, "", http.StatusInternalServerError)
	}

	if len(users) > 1 {
		return nil, model.NewAppError("SqlUserStore.GetForLogin", "store.sql_user.get_for_login.multiple_users", nil, "", http.StatusInternalServerError)
	}

	return users[0], nil
}

func (us SqlUserStore) GetProfileByIds(userIds []string, options *store.UserGetByIdsOpts, _ bool) ([]*model.User, *model.AppError) {
	if options == nil {
		options = &store.UserGetByIdsOpts{}
	}

	users := []*model.User{}
	query := us.usersQuery.
		Where(map[string]interface{}{
			"u.Id": userIds,
		}).
		OrderBy("u.Username ASC")

	if options.Since > 0 {
		query = query.Where(sq.Gt(map[string]interface{}{
			"u.UpdateAt": options.Since,
		}))
	}

	query = applyViewRestrictionsFilter(query, options.ViewRestrictions, true)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfileByIds", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetProfileByIds", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (us SqlUserStore) GetChannelGroupUsers(channelID string) ([]*model.User, *model.AppError) {
	query := applyChannelGroupConstrainedFilter(us.usersQuery, channelID)

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.GetChannelGroupUsers", "store.sql_user.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var users []*model.User
	if _, err := us.GetReplica().Select(&users, queryString, args...); err != nil {
		return nil, model.NewAppError("SqlUserStore.GetChannelGroupUsers", "store.sql_user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	for _, u := range users {
		u.Sanitize(map[string]bool{})
	}

	return users, nil
}

func (us SqlUserStore) UpdateUpdateAt(userId string) (int64, *model.AppError) {
	curTime := model.GetMillis()

	if _, err := us.GetMaster().Exec("UPDATE Users SET UpdateAt = :Time WHERE Id = :UserId", map[string]interface{}{"Time": curTime, "UserId": userId}); err != nil {
		return curTime, model.NewAppError("SqlUserStore.UpdateUpdateAt", "store.sql_user.update_update.app_error", nil, "user_id="+userId, http.StatusInternalServerError)
	}

	return curTime, nil
}
