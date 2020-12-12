package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"

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
	*SqlStore
}

func newSqlGroupStore(sqlStore *SqlStore) store.GroupStore {
	s := &SqlGroupStore{
		SqlStore: sqlStore,
	}
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

func (s *SqlGroupStore) Create(group *model.Group) (*model.Group, error) {
	if len(group.Id) != 0 {
		return nil, store.NewErrInvalidInput("Group", "id", group.Id)
	}

	if err := group.IsValidForCreate(); err != nil {
		return nil, err
	}

	group.Id = model.NewId()
	group.CreateAt = model.GetMillis()
	group.UpdateAt = group.CreateAt

	if err := s.GetMaster().Insert(group); err != nil {
		if IsUniqueConstraintError(err, []string{"Name", "groups_name_key"}) {
			return nil, errors.Wrapf(err, "Group with name %s already exists", *group.Name)
		}
		return nil, errors.Wrap(err, "failed to save Group")
	}

	return group, nil
}

func (s *SqlGroupStore) Get(groupId string) (*model.Group, error) {
	var group *model.Group
	if err := s.GetReplica().SelectOne(&group, "SELECT * from UserGroups WHERE Id = :Id", map[string]interface{}{"Id": groupId}); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Group", groupId)
		}
		return nil, errors.Wrapf(err, "failed to get Group with id=%s", groupId)
	}

	return group, nil
}

func (s *SqlGroupStore) GetByName(name string, opts model.GroupSearchOpts) (*model.Group, error) {
	var group *model.Group
	query := s.getQueryBuilder().Select("*").From("UserGroups").Where(sq.Eq{"Name": name})
	if opts.FilterAllowReference {
		query = query.Where("AllowReference = true")
	}

	queryString, args, err := query.ToSql()

	if err != nil {
		return nil, errors.Wrap(err, "get_by_name_tosql")
	}
	if err := s.GetReplica().SelectOne(&group, queryString, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Group", fmt.Sprintf("name=%s", name))
		}
		return nil, errors.Wrapf(err, "failed to get Group with name=%s", name)
	}

	return group, nil
}

func (s *SqlGroupStore) UpsertMember(groupID string, userID string) (*model.GroupMember, error) {
	member := &model.GroupMember{
		GroupId:  groupID,
		UserId:   userID,
		CreateAt: model.GetMillis(),
	}

	if err := member.IsValid(); err != nil {
		return nil, err
	}

	var retrievedGroup *model.Group
	if err := s.GetReplica().SelectOne(&retrievedGroup, "SELECT * FROM UserGroups WHERE Id = :Id", map[string]interface{}{"Id": groupID}); err != nil {
		return nil, errors.Wrapf(err, "failed to get UserGroup with groupId=%s and userId=%s", groupID, userID)
	}

	var retrievedMember *model.GroupMember
	if err := s.GetReplica().SelectOne(&retrievedMember, "SELECT * FROM GroupMembers WHERE GroupId = :GroupId AND UserId = :UserId", map[string]interface{}{"GroupId": member.GroupId, "UserId": member.UserId}); err != nil {
		if err != sql.ErrNoRows {
			return nil, errors.Wrapf(err, "failed to get GroupMember with groupId=%s and userId=%s", groupID, userID)
		}
	}

	if retrievedMember == nil {
		if err := s.GetMaster().Insert(member); err != nil {
			if IsUniqueConstraintError(err, []string{"GroupId", "UserId", "groupmembers_pkey", "PRIMARY"}) {
				return nil, store.NewErrInvalidInput("Member", "<groupId, userId>", fmt.Sprintf("<%s, %s>", groupID, userID))
			}
			return nil, errors.Wrap(err, "failed to save Member")
		}
	} else {
		member.DeleteAt = 0
		var rowsChanged int64
		var err error
		if rowsChanged, err = s.GetMaster().Update(member); err != nil {
			return nil, errors.Wrapf(err, "failed to update GroupMember with groupId=%s and userId=%s", groupID, userID)
		}
		if rowsChanged > 1 {
			return nil, errors.Wrapf(err, "multiple GroupMembers were updated: %d", rowsChanged)
		}
	}

	return member, nil
}

func groupSyncableToGroupTeam(groupSyncable *model.GroupSyncable) *groupTeam {
	return &groupTeam{
		GroupSyncable: *groupSyncable,
		TeamId:        groupSyncable.SyncableId,
	}
}

func groupSyncableToGroupChannel(groupSyncable *model.GroupSyncable) *groupChannel {
	return &groupChannel{
		GroupSyncable: *groupSyncable,
		ChannelId:     groupSyncable.SyncableId,
	}
}

func (s *SqlGroupStore) CreateGroupSyncable(groupSyncable *model.GroupSyncable) (*model.GroupSyncable, error) {
	if err := groupSyncable.IsValid(); err != nil {
		return nil, err
	}

	// Reset values that shouldn't be updatable by parameter
	groupSyncable.DeleteAt = 0
	groupSyncable.CreateAt = model.GetMillis()
	groupSyncable.UpdateAt = groupSyncable.CreateAt

	var insertErr error

	switch groupSyncable.Type {
	case model.GroupSyncableTypeTeam:
		if _, err := s.Team().Get(groupSyncable.SyncableId); err != nil {
			return nil, err
		}

		insertErr = s.GetMaster().Insert(groupSyncableToGroupTeam(groupSyncable))
	case model.GroupSyncableTypeChannel:
		if _, err := s.Channel().Get(groupSyncable.SyncableId, false); err != nil {
			return nil, err
		}

		insertErr = s.GetMaster().Insert(groupSyncableToGroupChannel(groupSyncable))
	default:
		return nil, fmt.Errorf("invalid GroupSyncableType: %s", groupSyncable.Type)
	}

	if insertErr != nil {
		return nil, errors.Wrap(insertErr, "unable to insert GroupSyncable")
	}

	return groupSyncable, nil
}

func (s *SqlGroupStore) GetGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, error) {
	groupSyncable, err := s.getGroupSyncable(groupID, syncableID, syncableType)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("GroupSyncable", fmt.Sprintf("groupId=%s, syncableId=%s, syncableType=%s", groupID, syncableID, syncableType))
		}
		return nil, errors.Wrapf(err, "failed to find GroupSyncable with groupId=%s, syncableId=%s, syncableType=%s", groupID, syncableID, syncableType)
	}

	return groupSyncable, nil
}

func (s *SqlGroupStore) getGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, error) {
	var err error
	var result interface{}

	switch syncableType {
	case model.GroupSyncableTypeTeam:
		result, err = s.GetReplica().Get(groupTeam{}, groupID, syncableID)
	case model.GroupSyncableTypeChannel:
		result, err = s.GetReplica().Get(groupChannel{}, groupID, syncableID)
	}

	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, sql.ErrNoRows
	}

	groupSyncable := model.GroupSyncable{}
	switch syncableType {
	case model.GroupSyncableTypeTeam:
		groupTeam := result.(*groupTeam)
		groupSyncable.SyncableId = groupTeam.TeamId
		groupSyncable.GroupId = groupTeam.GroupId
		groupSyncable.AutoAdd = groupTeam.AutoAdd
		groupSyncable.CreateAt = groupTeam.CreateAt
		groupSyncable.DeleteAt = groupTeam.DeleteAt
		groupSyncable.UpdateAt = groupTeam.UpdateAt
		groupSyncable.Type = syncableType
	case model.GroupSyncableTypeChannel:
		groupChannel := result.(*groupChannel)
		groupSyncable.SyncableId = groupChannel.ChannelId
		groupSyncable.GroupId = groupChannel.GroupId
		groupSyncable.AutoAdd = groupChannel.AutoAdd
		groupSyncable.CreateAt = groupChannel.CreateAt
		groupSyncable.DeleteAt = groupChannel.DeleteAt
		groupSyncable.UpdateAt = groupChannel.UpdateAt
		groupSyncable.Type = syncableType
	default:
		return nil, fmt.Errorf("unable to convert syncableType: %s", syncableType.String())
	}

	return &groupSyncable, nil
}

func (s *SqlGroupStore) AdminRoleGroupsForSyncableMember(userID, syncableID string, syncableType model.GroupSyncableType) ([]string, error) {
	var groupIds []string

	query := fmt.Sprintf(`
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

	_, err := s.GetReplica().Select(&groupIds, query, map[string]interface{}{"UserId": userID, fmt.Sprintf("%sId", syncableType): syncableID})
	if err != nil {
		return nil, errors.Wrap(err, "failed to find Group ids")
	}

	return groupIds, nil
}
