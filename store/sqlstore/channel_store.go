package sqlstore

import (
	"net/http"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
	"github.com/masterhung0112/go_server/store"
	"strings"
	"time"
	"database/sql"
  "github.com/masterhung0112/go_server/model"
  sq "github.com/Masterminds/squirrel"
)


const (
	ALL_CHANNEL_MEMBERS_FOR_USER_CACHE_SIZE     = model.SESSION_CACHE_SIZE
	ALL_CHANNEL_MEMBERS_FOR_USER_CACHE_DURATION = 15 * time.Minute // 15 mins

	ALL_CHANNEL_MEMBERS_NOTIFY_PROPS_FOR_CHANNEL_CACHE_SIZE     = model.SESSION_CACHE_SIZE
	ALL_CHANNEL_MEMBERS_NOTIFY_PROPS_FOR_CHANNEL_CACHE_DURATION = 30 * time.Minute // 30 mins

	CHANNEL_CACHE_DURATION = 15 * time.Minute // 15 mins
)

type SqlChannelStore struct {
	SqlStore
	// metrics einterfaces.MetricsInterface
}

type channelMember struct {
	ChannelId    string
	UserId       string
	Roles        string
	LastViewedAt int64
	MsgCount     int64
	MentionCount int64
	NotifyProps  model.StringMap
	LastUpdateAt int64
	SchemeUser   sql.NullBool
	SchemeAdmin  sql.NullBool
	SchemeGuest  sql.NullBool
}

func NewChannelMemberFromModel(cm *model.ChannelMember) *channelMember {
	return &channelMember{
		ChannelId:    cm.ChannelId,
		UserId:       cm.UserId,
		Roles:        cm.ExplicitRoles,
		LastViewedAt: cm.LastViewedAt,
		MsgCount:     cm.MsgCount,
		MentionCount: cm.MentionCount,
		NotifyProps:  cm.NotifyProps,
		LastUpdateAt: cm.LastUpdateAt,
		SchemeGuest:  sql.NullBool{Valid: true, Bool: cm.SchemeGuest},
		SchemeUser:   sql.NullBool{Valid: true, Bool: cm.SchemeUser},
		SchemeAdmin:  sql.NullBool{Valid: true, Bool: cm.SchemeAdmin},
	}
}

type channelMemberWithSchemeRoles struct {
	ChannelId                     string
	UserId                        string
	Roles                         string
	LastViewedAt                  int64
	MsgCount                      int64
	MentionCount                  int64
	NotifyProps                   model.StringMap
	LastUpdateAt                  int64
	SchemeGuest                   sql.NullBool
	SchemeUser                    sql.NullBool
	SchemeAdmin                   sql.NullBool
	TeamSchemeDefaultGuestRole    sql.NullString
	TeamSchemeDefaultUserRole     sql.NullString
	TeamSchemeDefaultAdminRole    sql.NullString
	ChannelSchemeDefaultGuestRole sql.NullString
	ChannelSchemeDefaultUserRole  sql.NullString
	ChannelSchemeDefaultAdminRole sql.NullString
}

func channelMemberSliceColumns() []string {
	return []string{"ChannelId", "UserId", "Roles", "LastViewedAt", "MsgCount", "MentionCount", "NotifyProps", "LastUpdateAt", "SchemeUser", "SchemeAdmin", "SchemeGuest"}
}

func channelMemberToSlice(member *model.ChannelMember) []interface{} {
	resultSlice := []interface{}{}
	resultSlice = append(resultSlice, member.ChannelId)
	resultSlice = append(resultSlice, member.UserId)
	resultSlice = append(resultSlice, member.ExplicitRoles)
	resultSlice = append(resultSlice, member.LastViewedAt)
	resultSlice = append(resultSlice, member.MsgCount)
	resultSlice = append(resultSlice, member.MentionCount)
	resultSlice = append(resultSlice, model.MapToJson(member.NotifyProps))
	resultSlice = append(resultSlice, member.LastUpdateAt)
	resultSlice = append(resultSlice, member.SchemeUser)
	resultSlice = append(resultSlice, member.SchemeAdmin)
	resultSlice = append(resultSlice, member.SchemeGuest)
	return resultSlice
}

type channelMemberWithSchemeRolesList []channelMemberWithSchemeRoles

func getChannelRoles(schemeGuest, schemeUser, schemeAdmin bool, defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole, defaultChannelGuestRole, defaultChannelUserRole, defaultChannelAdminRole string, roles []string) rolesInfo {
	result := rolesInfo{
		roles:         []string{},
		explicitRoles: []string{},
		schemeGuest:   schemeGuest,
		schemeUser:    schemeUser,
		schemeAdmin:   schemeAdmin,
	}

	// Identify any scheme derived roles that are in "Roles" field due to not yet being migrated, and exclude
	// them from ExplicitRoles field.
	for _, role := range roles {
		switch role {
		case model.CHANNEL_GUEST_ROLE_ID:
			result.schemeGuest = true
		case model.CHANNEL_USER_ROLE_ID:
			result.schemeUser = true
		case model.CHANNEL_ADMIN_ROLE_ID:
			result.schemeAdmin = true
		default:
			result.explicitRoles = append(result.explicitRoles, role)
			result.roles = append(result.roles, role)
		}
	}

	// Add any scheme derived roles that are not in the Roles field due to being Implicit from the Scheme, and add
	// them to the Roles field for backwards compatibility reasons.
	var schemeImpliedRoles []string
	if result.schemeGuest {
		if defaultChannelGuestRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultChannelGuestRole)
		} else if defaultTeamGuestRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamGuestRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_GUEST_ROLE_ID)
		}
	}
	if result.schemeUser {
		if defaultChannelUserRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultChannelUserRole)
		} else if defaultTeamUserRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamUserRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_USER_ROLE_ID)
		}
	}
	if result.schemeAdmin {
		if defaultChannelAdminRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultChannelAdminRole)
		} else if defaultTeamAdminRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamAdminRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_ADMIN_ROLE_ID)
		}
	}
	for _, impliedRole := range schemeImpliedRoles {
		alreadyThere := false
		for _, role := range result.roles {
			if role == impliedRole {
				alreadyThere = true
				break
			}
		}
		if !alreadyThere {
			result.roles = append(result.roles, impliedRole)
		}
	}
	return result
}

func (db channelMemberWithSchemeRoles) ToModel() *model.ChannelMember {
	// Identify any system-wide scheme derived roles that are in "Roles" field due to not yet being migrated,
	// and exclude them from ExplicitRoles field.
	schemeGuest := db.SchemeGuest.Valid && db.SchemeGuest.Bool
	schemeUser := db.SchemeUser.Valid && db.SchemeUser.Bool
	schemeAdmin := db.SchemeAdmin.Valid && db.SchemeAdmin.Bool

	defaultTeamGuestRole := ""
	if db.TeamSchemeDefaultGuestRole.Valid {
		defaultTeamGuestRole = db.TeamSchemeDefaultGuestRole.String
	}

	defaultTeamUserRole := ""
	if db.TeamSchemeDefaultUserRole.Valid {
		defaultTeamUserRole = db.TeamSchemeDefaultUserRole.String
	}

	defaultTeamAdminRole := ""
	if db.TeamSchemeDefaultAdminRole.Valid {
		defaultTeamAdminRole = db.TeamSchemeDefaultAdminRole.String
	}

	defaultChannelGuestRole := ""
	if db.ChannelSchemeDefaultGuestRole.Valid {
		defaultChannelGuestRole = db.ChannelSchemeDefaultGuestRole.String
	}

	defaultChannelUserRole := ""
	if db.ChannelSchemeDefaultUserRole.Valid {
		defaultChannelUserRole = db.ChannelSchemeDefaultUserRole.String
	}

	defaultChannelAdminRole := ""
	if db.ChannelSchemeDefaultAdminRole.Valid {
		defaultChannelAdminRole = db.ChannelSchemeDefaultAdminRole.String
	}

	rolesResult := getChannelRoles(
		schemeGuest, schemeUser, schemeAdmin,
		defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole,
		defaultChannelGuestRole, defaultChannelUserRole, defaultChannelAdminRole,
		strings.Fields(db.Roles),
	)
	return &model.ChannelMember{
		ChannelId:     db.ChannelId,
		UserId:        db.UserId,
		Roles:         strings.Join(rolesResult.roles, " "),
		LastViewedAt:  db.LastViewedAt,
		MsgCount:      db.MsgCount,
		MentionCount:  db.MentionCount,
		NotifyProps:   db.NotifyProps,
		LastUpdateAt:  db.LastUpdateAt,
		SchemeAdmin:   rolesResult.schemeAdmin,
		SchemeUser:    rolesResult.schemeUser,
		SchemeGuest:   rolesResult.schemeGuest,
		ExplicitRoles: strings.Join(rolesResult.explicitRoles, " "),
	}
}

func (db channelMemberWithSchemeRolesList) ToModel() *model.ChannelMembers {
	cms := model.ChannelMembers{}

	for _, cm := range db {
		cms = append(cms, *cm.ToModel())
	}

	return &cms
}

type allChannelMember struct {
	ChannelId                     string
	Roles                         string
	SchemeGuest                   sql.NullBool
	SchemeUser                    sql.NullBool
	SchemeAdmin                   sql.NullBool
	TeamSchemeDefaultGuestRole    sql.NullString
	TeamSchemeDefaultUserRole     sql.NullString
	TeamSchemeDefaultAdminRole    sql.NullString
	ChannelSchemeDefaultGuestRole sql.NullString
	ChannelSchemeDefaultUserRole  sql.NullString
	ChannelSchemeDefaultAdminRole sql.NullString
}

type allChannelMembers []allChannelMember

func (db allChannelMember) Process() (string, string) {
	roles := strings.Fields(db.Roles)

	// Add any scheme derived roles that are not in the Roles field due to being Implicit from the Scheme, and add
	// them to the Roles field for backwards compatibility reasons.
	var schemeImpliedRoles []string
	if db.SchemeGuest.Valid && db.SchemeGuest.Bool {
		if db.ChannelSchemeDefaultGuestRole.Valid && db.ChannelSchemeDefaultGuestRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.ChannelSchemeDefaultGuestRole.String)
		} else if db.TeamSchemeDefaultGuestRole.Valid && db.TeamSchemeDefaultGuestRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.TeamSchemeDefaultGuestRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_GUEST_ROLE_ID)
		}
	}
	if db.SchemeUser.Valid && db.SchemeUser.Bool {
		if db.ChannelSchemeDefaultUserRole.Valid && db.ChannelSchemeDefaultUserRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.ChannelSchemeDefaultUserRole.String)
		} else if db.TeamSchemeDefaultUserRole.Valid && db.TeamSchemeDefaultUserRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.TeamSchemeDefaultUserRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_USER_ROLE_ID)
		}
	}
	if db.SchemeAdmin.Valid && db.SchemeAdmin.Bool {
		if db.ChannelSchemeDefaultAdminRole.Valid && db.ChannelSchemeDefaultAdminRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.ChannelSchemeDefaultAdminRole.String)
		} else if db.TeamSchemeDefaultAdminRole.Valid && db.TeamSchemeDefaultAdminRole.String != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, db.TeamSchemeDefaultAdminRole.String)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.CHANNEL_ADMIN_ROLE_ID)
		}
	}
	for _, impliedRole := range schemeImpliedRoles {
		alreadyThere := false
		for _, role := range roles {
			if role == impliedRole {
				alreadyThere = true
			}
		}
		if !alreadyThere {
			roles = append(roles, impliedRole)
		}
	}

	return db.ChannelId, strings.Join(roles, " ")
}

func (db allChannelMembers) ToMapStringString() map[string]string {
	result := make(map[string]string)

	for _, item := range db {
		key, value := item.Process()
		result[key] = value
	}

	return result
}

// publicChannel is a subset of the metadata corresponding to public channels only.
type publicChannel struct {
	Id          string `json:"id"`
	DeleteAt    int64  `json:"delete_at"`
	TeamId      string `json:"team_id"`
	DisplayName string `json:"display_name"`
	Name        string `json:"name"`
	Header      string `json:"header"`
	Purpose     string `json:"purpose"`
}

//TODO:Open this
// var allChannelMembersForUserCache = cache.NewLRU(&cache.LRUOptions{
// 	Size: ALL_CHANNEL_MEMBERS_FOR_USER_CACHE_SIZE,
// })
// var allChannelMembersNotifyPropsForChannelCache = cache.NewLRU(&cache.LRUOptions{
// 	Size: ALL_CHANNEL_MEMBERS_NOTIFY_PROPS_FOR_CHANNEL_CACHE_SIZE,
// })
// var channelByNameCache = cache.NewLRU(&cache.LRUOptions{
// 	Size: model.CHANNEL_CACHE_SIZE,
// })

// func newSqlChannelStore(sqlStore SqlStore, metrics einterfaces.MetricsInterface) store.ChannelStore {
func newSqlChannelStore(sqlStore SqlStore) store.ChannelStore {

	s := &SqlChannelStore{
		SqlStore: sqlStore,
		// metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Channel{}, "Channels").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("TeamId").SetMaxSize(26)
		table.ColMap("Type").SetMaxSize(1)
		table.ColMap("DisplayName").SetMaxSize(64)
		table.ColMap("Name").SetMaxSize(64)
		table.SetUniqueTogether("Name", "TeamId")
		table.ColMap("Header").SetMaxSize(1024)
		table.ColMap("Purpose").SetMaxSize(250)
		table.ColMap("CreatorId").SetMaxSize(26)
		table.ColMap("SchemeId").SetMaxSize(26)

		tablem := db.AddTableWithName(channelMember{}, "ChannelMembers").SetKeys(false, "ChannelId", "UserId")
		tablem.ColMap("ChannelId").SetMaxSize(26)
		tablem.ColMap("UserId").SetMaxSize(26)
		tablem.ColMap("Roles").SetMaxSize(64)
		tablem.ColMap("NotifyProps").SetMaxSize(2000)

		tablePublicChannels := db.AddTableWithName(publicChannel{}, "PublicChannels").SetKeys(false, "Id")
		tablePublicChannels.ColMap("Id").SetMaxSize(26)
		tablePublicChannels.ColMap("TeamId").SetMaxSize(26)
		tablePublicChannels.ColMap("DisplayName").SetMaxSize(64)
		tablePublicChannels.ColMap("Name").SetMaxSize(64)
		tablePublicChannels.SetUniqueTogether("Name", "TeamId")
		tablePublicChannels.ColMap("Header").SetMaxSize(1024)
		tablePublicChannels.ColMap("Purpose").SetMaxSize(250)

		tableSidebarCategories := db.AddTableWithName(model.SidebarCategory{}, "SidebarCategories").SetKeys(false, "Id")
		tableSidebarCategories.ColMap("Id").SetMaxSize(26)
		tableSidebarCategories.ColMap("UserId").SetMaxSize(26)
		tableSidebarCategories.ColMap("TeamId").SetMaxSize(26)
		tableSidebarCategories.ColMap("Sorting").SetMaxSize(64)
		tableSidebarCategories.ColMap("Type").SetMaxSize(64)
		tableSidebarCategories.ColMap("DisplayName").SetMaxSize(64)

		tableSidebarChannels := db.AddTableWithName(model.SidebarChannel{}, "SidebarChannels").SetKeys(false, "ChannelId", "UserId", "CategoryId")
		tableSidebarChannels.ColMap("ChannelId").SetMaxSize(26)
		tableSidebarChannels.ColMap("UserId").SetMaxSize(26)
		tableSidebarChannels.ColMap("CategoryId").SetMaxSize(26)
	}

	return s
}


// Save writes the (non-direct) channel channel to the database.
func (s SqlChannelStore) Save(channel *model.Channel, maxChannelsPerTeam int64) (*model.Channel, error) {
	if channel.DeleteAt != 0 {
		return nil, store.NewErrInvalidInput("Channel", "DeleteAt", channel.DeleteAt)
	}

	if channel.Type == model.CHANNEL_DIRECT {
		return nil, store.NewErrInvalidInput("Channel", "Type", channel.Type)
	}

	var newChannel *model.Channel
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return nil, errors.Wrap(err, "begin_transaction")
	}
	defer finalizeTransaction(transaction)

	newChannel, err = s.saveChannelT(transaction, channel, maxChannelsPerTeam)
	if err != nil {
		return newChannel, err
	}

	// Additionally propagate the write to the PublicChannels table.
	if err = s.upsertPublicChannelT(transaction, newChannel); err != nil {
		return nil, errors.Wrap(err, "upsert_public_channel")
	}

	if err = transaction.Commit(); err != nil {
		return nil, errors.Wrap(err, "commit_transaction")
	}
	// There are cases when in case of conflict, the original channel value is returned.
	// So we return both and let the caller do the checks.
	return newChannel, err
}

func (s SqlChannelStore) upsertPublicChannelT(transaction *gorp.Transaction, channel *model.Channel) error {
	publicChannel := &publicChannel{
		Id:          channel.Id,
		DeleteAt:    channel.DeleteAt,
		TeamId:      channel.TeamId,
		DisplayName: channel.DisplayName,
		Name:        channel.Name,
		Header:      channel.Header,
		Purpose:     channel.Purpose,
	}

	if channel.Type != model.CHANNEL_OPEN {
		if _, err := transaction.Delete(publicChannel); err != nil {
			return errors.Wrap(err, "failed to delete public channel")
		}

		return nil
	}

	if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		// Leverage native upsert for MySQL, since RowsAffected returns 0 if the row exists
		// but no changes were made, breaking the update-then-insert paradigm below when
		// the row already exists. (Postgres 9.4 doesn't support native upsert.)
		if _, err := transaction.Exec(`
			INSERT INTO
			    PublicChannels(Id, DeleteAt, TeamId, DisplayName, Name, Header, Purpose)
			VALUES
			    (:Id, :DeleteAt, :TeamId, :DisplayName, :Name, :Header, :Purpose)
			ON DUPLICATE KEY UPDATE
			    DeleteAt = :DeleteAt,
			    TeamId = :TeamId,
			    DisplayName = :DisplayName,
			    Name = :Name,
			    Header = :Header,
			    Purpose = :Purpose;
		`, map[string]interface{}{
			"Id":          publicChannel.Id,
			"DeleteAt":    publicChannel.DeleteAt,
			"TeamId":      publicChannel.TeamId,
			"DisplayName": publicChannel.DisplayName,
			"Name":        publicChannel.Name,
			"Header":      publicChannel.Header,
			"Purpose":     publicChannel.Purpose,
		}); err != nil {
			return errors.Wrap(err, "failed to insert public channel")
		}
	} else {
		count, err := transaction.Update(publicChannel)
		if err != nil {
			return errors.Wrap(err, "failed to update public channel")
		}
		if count > 0 {
			return nil
		}

		if err := transaction.Insert(publicChannel); err != nil {
			return errors.Wrap(err, "failed to insert public channel")
		}
	}

	return nil
}

func (s SqlChannelStore) saveChannelT(transaction *gorp.Transaction, channel *model.Channel, maxChannelsPerTeam int64) (*model.Channel, error) {
	if len(channel.Id) > 0 {
		return nil, store.NewErrInvalidInput("Channel", "Id", channel.Id)
	}

	channel.PreSave()
	if err := channel.IsValid(); err != nil { // TODO: this needs to return plain error in v6.
		return nil, err // we just pass through the error as-is for now.
	}

	if channel.Type != model.CHANNEL_DIRECT && channel.Type != model.CHANNEL_GROUP && maxChannelsPerTeam >= 0 {
		if count, err := transaction.SelectInt("SELECT COUNT(0) FROM Channels WHERE TeamId = :TeamId AND DeleteAt = 0 AND (Type = 'O' OR Type = 'P')", map[string]interface{}{"TeamId": channel.TeamId}); err != nil {
			return nil, errors.Wrapf(err, "save_channel_count: teamId=%s", channel.TeamId)
		} else if count >= maxChannelsPerTeam {
			return nil, store.NewErrLimitExceeded("channels_per_team", int(count), "teamId="+channel.TeamId)
		}
	}

	if err := transaction.Insert(channel); err != nil {
		if IsUniqueConstraintError(err, []string{"Name", "channels_name_teamid_key"}) {
			dupChannel := model.Channel{}
			s.GetMaster().SelectOne(&dupChannel, "SELECT * FROM Channels WHERE TeamId = :TeamId AND Name = :Name", map[string]interface{}{"TeamId": channel.TeamId, "Name": channel.Name})
			return &dupChannel, store.NewErrConflict("Channel", err, "id="+channel.Id)
		}
		return nil, errors.Wrapf(err, "save_channel: id=%s", channel.Id)
	}
	return channel, nil
}

func (s SqlChannelStore) GetAllChannelMembersForUser(userId string, allowFromCache bool, includeDeleted bool) (map[string]string, *model.AppError) {
	cache_key := userId
	if includeDeleted {
		cache_key += "_deleted"
  }

  //TODO: Open this
	// if allowFromCache {
  //   var ids map[string]string

	// 	if err := allChannelMembersForUserCache.Get(cache_key, &ids); err == nil {
	// 		// if s.metrics != nil {
	// 		// 	s.metrics.IncrementMemCacheHitCounter("All Channel Members for User")
	// 		// }
	// 		return ids, nil
	// 	}
	// }

  //TODO: Open this
	// if s.metrics != nil {
	// 	s.metrics.IncrementMemCacheMissCounter("All Channel Members for User")
	// }

	failure := func(err error) *model.AppError {
		// TODO: This error key would go away once this store method is migrated to return plain errors
		return model.NewAppError(
			"SqlChannelStore.GetAllChannelMembersForUser",
			"app.channel.get_channels.get.app_error",
			nil,
			"userId="+userId+", err="+err.Error(),
			http.StatusInternalServerError,
		)
	}

	query := s.getQueryBuilder().
		Select(`
				ChannelMembers.ChannelId, ChannelMembers.Roles, ChannelMembers.SchemeGuest,
				ChannelMembers.SchemeUser, ChannelMembers.SchemeAdmin,
				TeamScheme.DefaultChannelGuestRole TeamSchemeDefaultGuestRole,
				TeamScheme.DefaultChannelUserRole TeamSchemeDefaultUserRole,
				TeamScheme.DefaultChannelAdminRole TeamSchemeDefaultAdminRole,
				ChannelScheme.DefaultChannelGuestRole ChannelSchemeDefaultGuestRole,
				ChannelScheme.DefaultChannelUserRole ChannelSchemeDefaultUserRole,
				ChannelScheme.DefaultChannelAdminRole ChannelSchemeDefaultAdminRole
		`).
		From("ChannelMembers").
		Join("Channels ON ChannelMembers.ChannelId = Channels.Id").
		LeftJoin("Schemes ChannelScheme ON Channels.SchemeId = ChannelScheme.Id").
		LeftJoin("Teams ON Channels.TeamId = Teams.Id").
		LeftJoin("Schemes TeamScheme ON Teams.SchemeId = TeamScheme.Id").
		Where(sq.Eq{"ChannelMembers.UserId": userId})
	if !includeDeleted {
		query = query.Where(sq.Eq{"Channels.DeleteAt": 0})
	}
	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, failure(err)
	}

	rows, err := s.GetReplica().Db.Query(queryString, args...)
	if err != nil {
		return nil, failure(err)
	}

	var data allChannelMembers
	defer rows.Close()
	for rows.Next() {
		var cm allChannelMember
		err = rows.Scan(
			&cm.ChannelId, &cm.Roles, &cm.SchemeGuest, &cm.SchemeUser,
			&cm.SchemeAdmin, &cm.TeamSchemeDefaultGuestRole, &cm.TeamSchemeDefaultUserRole,
			&cm.TeamSchemeDefaultAdminRole, &cm.ChannelSchemeDefaultGuestRole,
			&cm.ChannelSchemeDefaultUserRole, &cm.ChannelSchemeDefaultAdminRole,
		)
		if err != nil {
			return nil, failure(err)
		}
		data = append(data, cm)
	}
	if err = rows.Err(); err != nil {
		return nil, failure(err)
	}
	ids := data.ToMapStringString()

  //TODO: Open this
	// if allowFromCache {
	// 	allChannelMembersForUserCache.SetWithExpiry(cache_key, ids, ALL_CHANNEL_MEMBERS_FOR_USER_CACHE_DURATION)
	// }
	return ids, nil
}


func (s SqlChannelStore) get(id string, master bool, allowFromCache bool) (*model.Channel, error) {
	var db *gorp.DbMap

	if master {
		db = s.GetMaster()
	} else {
		db = s.GetReplica()
	}

	obj, err := db.Get(model.Channel{}, id)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find channel with id = %s", id)
	}

	if obj == nil {
		return nil, store.NewErrNotFound("Channel", id)
	}

	ch := obj.(*model.Channel)
	return ch, nil
}

func (s SqlChannelStore) Get(id string, allowFromCache bool) (*model.Channel, error) {
	return s.get(id, false, allowFromCache)
}

func (s SqlChannelStore) GetMemberForPost(postId string, userId string) (*model.ChannelMember, *model.AppError) {
	var dbMember channelMemberWithSchemeRoles
	query := `
		SELECT
			ChannelMembers.*,
			TeamScheme.DefaultChannelGuestRole TeamSchemeDefaultGuestRole,
			TeamScheme.DefaultChannelUserRole TeamSchemeDefaultUserRole,
			TeamScheme.DefaultChannelAdminRole TeamSchemeDefaultAdminRole,
			ChannelScheme.DefaultChannelGuestRole ChannelSchemeDefaultGuestRole,
			ChannelScheme.DefaultChannelUserRole ChannelSchemeDefaultUserRole,
			ChannelScheme.DefaultChannelAdminRole ChannelSchemeDefaultAdminRole
		FROM
			ChannelMembers
		INNER JOIN
			Posts ON ChannelMembers.ChannelId = Posts.ChannelId
		INNER JOIN
			Channels ON ChannelMembers.ChannelId = Channels.Id
		LEFT JOIN
			Schemes ChannelScheme ON Channels.SchemeId = ChannelScheme.Id
		LEFT JOIN
			Teams ON Channels.TeamId = Teams.Id
		LEFT JOIN
			Schemes TeamScheme ON Teams.SchemeId = TeamScheme.Id
		WHERE
			ChannelMembers.UserId = :UserId
		AND
			Posts.Id = :PostId`
	if err := s.GetReplica().SelectOne(&dbMember, query, map[string]interface{}{"UserId": userId, "PostId": postId}); err != nil {
		return nil, model.NewAppError("SqlChannelStore.GetMemberForPost", "store.sql_channel.get_member_for_post.app_error", nil, "postId="+postId+", err="+err.Error(), http.StatusInternalServerError)
	}
	return dbMember.ToModel(), nil
}

func (s SqlChannelStore) GetForPost(postId string) (*model.Channel, error) {
	channel := &model.Channel{}
	if err := s.GetReplica().SelectOne(
		channel,
		`SELECT
			Channels.*
		FROM
			Channels,
			Posts
		WHERE
			Channels.Id = Posts.ChannelId
			AND Posts.Id = :PostId`, map[string]interface{}{"PostId": postId}); err != nil {
		return nil, errors.Wrapf(err, "failed to get Channel with postId=%s", postId)

	}
	return channel, nil
}