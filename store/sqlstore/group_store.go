package sqlstore

import (
	"database/sql"
	"fmt"
	"net/http"

	sq "github.com/Masterminds/squirrel"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
)

type selectType int

const (
	selectGroups selectType = iota
	selectCountGroups
)

type groupTeam struct {
	model.GroupSyncable
	TeamId string `db:"TeamId"`
}

type groupChannel struct {
	model.GroupSyncable
	ChannelId string `db:"ChannelId"`
}

type groupTeamJoin struct {
	groupTeam
	TeamDisplayName string `db:"TeamDisplayName"`
	TeamType        string `db:"TeamType"`
}

type groupChannelJoin struct {
	groupChannel
	ChannelDisplayName string `db:"ChannelDisplayName"`
	TeamDisplayName    string `db:"TeamDisplayName"`
	TeamType           string `db:"TeamType"`
	ChannelType        string `db:"ChannelType"`
	TeamID             string `db:"TeamId"`
}

type SqlGroupStore struct {
	SqlStore
}

func newSqlGroupStore(sqlStore SqlStore) store.GroupStore {
	s := &SqlGroupStore{SqlStore: sqlStore}
	for _, db := range sqlStore.GetAllConns() {
		groups := db.AddTableWithName(model.Group{}, "UserGroups").SetKeys(false, "Id")
		groups.ColMap("Id").SetMaxSize(26)
		groups.ColMap("Name").SetMaxSize(model.GroupNameMaxLength).SetUnique(true)
		groups.ColMap("DisplayName").SetMaxSize(model.GroupDisplayNameMaxLength)
		groups.ColMap("Description").SetMaxSize(model.GroupDescriptionMaxLength)
		groups.ColMap("Source").SetMaxSize(model.GroupSourceMaxLength)
		groups.ColMap("RemoteId").SetMaxSize(model.GroupRemoteIDMaxLength)
		groups.SetUniqueTogether("Source", "RemoteId")

		groupMembers := db.AddTableWithName(model.GroupMember{}, "GroupMembers").SetKeys(false, "GroupId", "UserId")
		groupMembers.ColMap("GroupId").SetMaxSize(26)
		groupMembers.ColMap("UserId").SetMaxSize(26)

		groupTeams := db.AddTableWithName(groupTeam{}, "GroupTeams").SetKeys(false, "GroupId", "TeamId")
		groupTeams.ColMap("GroupId").SetMaxSize(26)
		groupTeams.ColMap("TeamId").SetMaxSize(26)

		groupChannels := db.AddTableWithName(groupChannel{}, "GroupChannels").SetKeys(false, "GroupId", "ChannelId")
		groupChannels.ColMap("GroupId").SetMaxSize(26)
		groupChannels.ColMap("ChannelId").SetMaxSize(26)
	}
	return s
}

func (s *SqlGroupStore) Create(group *model.Group) (*model.Group, *model.AppError) {
	if len(group.Id) != 0 {
		return nil, model.NewAppError("SqlGroupStore.GroupCreate", "model.group.id.app_error", nil, "", http.StatusBadRequest)
	}

	if err := group.IsValidForCreate(); err != nil {
		return nil, err
	}

	group.Id = model.NewId()
	group.CreateAt = model.GetMillis()
	group.UpdateAt = group.CreateAt

	if err := s.GetMaster().Insert(group); err != nil {
		if IsUniqueConstraintError(err, []string{"Name", "groups_name_key"}) {
			return nil, model.NewAppError("SqlGroupStore.GroupCreate", "store.sql_group.unique_constraint", nil, err.Error(), http.StatusInternalServerError)
		}
		return nil, model.NewAppError("SqlGroupStore.GroupCreate", "store.insert_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return group, nil
}

func (s *SqlGroupStore) Get(groupId string) (*model.Group, *model.AppError) {
	var group *model.Group
	if err := s.GetReplica().SelectOne(&group, "SELECT * from UserGroups WHERE Id = :Id", map[string]interface{}{"Id": groupId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlGroupStore.GroupGet", "store.sql_group.no_rows", nil, err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlGroupStore.GroupGet", "store.select_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return group, nil
}

func (s *SqlGroupStore) GetByName(name string, opts model.GroupSearchOpts) (*model.Group, *model.AppError) {
	var group *model.Group
	query := s.getQueryBuilder().Select("*").From("UserGroups").Where(sq.Eq{"Name": name})
	if opts.FilterAllowReference {
		query = query.Where("AllowReference = true")
	}

	queryString, args, err := query.ToSql()

	if err != nil {
		return nil, model.NewAppError("SqlGroupStore.GetByName", "store.sql_group.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	if err := s.GetReplica().SelectOne(&group, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlGroupStore.GroupGetByName", "store.sql_group.no_rows", nil, err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlGroupStore.GroupGetByName", "store.select_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return group, nil
}

func (s *SqlGroupStore) AdminRoleGroupsForSyncableMember(userID, syncableID string, syncableType model.GroupSyncableType) ([]string, *model.AppError) {
	var groupIds []string

	sql := fmt.Sprintf(`
		SELECT
			GroupMembers.GroupId
		FROM
			GroupMembers
		INNER JOIN
			Group%[1]ss ON Group%[1]ss.GroupId = GroupMembers.GroupId
		WHERE
			GroupMembers.UserId = :UserId
			AND GroupMembers.DeleteAt = 0
			AND %[1]sId = :%[1]sId
			AND Group%[1]ss.DeleteAt = 0
			AND Group%[1]ss.SchemeAdmin = TRUE`, syncableType)

	_, err := s.GetReplica().Select(&groupIds, sql, map[string]interface{}{"UserId": userID, fmt.Sprintf("%sId", syncableType): syncableID})
	if err != nil {
		return nil, model.NewAppError("SqlGroupStore AdminRoleGroupsForSyncableMember", "store.select_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return groupIds, nil
}
