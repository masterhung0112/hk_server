package sqlstore

import (
	"database/sql"
	"fmt"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
	"net/http"
	"strings"

	sq "github.com/Masterminds/squirrel"
)

const (
	TEAM_MEMBER_EXISTS_ERROR = "store.sql_team.save_member.exists.app_error"
)

type SqlTeamStore struct {
	SqlStore

	teamsQuery sq.SelectBuilder
}

type teamMember struct {
	TeamId      string
	UserId      string
	Roles       string
	DeleteAt    int64
	SchemeUser  sql.NullBool
	SchemeAdmin sql.NullBool
	SchemeGuest sql.NullBool
}

func NewTeamMemberFromModel(tm *model.TeamMember) *teamMember {
	return &teamMember{
		TeamId:      tm.TeamId,
		UserId:      tm.UserId,
		Roles:       tm.ExplicitRoles,
		DeleteAt:    tm.DeleteAt,
		SchemeGuest: sql.NullBool{Valid: true, Bool: tm.SchemeGuest},
		SchemeUser:  sql.NullBool{Valid: true, Bool: tm.SchemeUser},
		SchemeAdmin: sql.NullBool{Valid: true, Bool: tm.SchemeAdmin},
	}
}

type teamMemberWithSchemeRoles struct {
	TeamId                     string
	UserId                     string
	Roles                      string
	DeleteAt                   int64
	SchemeGuest                sql.NullBool
	SchemeUser                 sql.NullBool
	SchemeAdmin                sql.NullBool
	TeamSchemeDefaultGuestRole sql.NullString
	TeamSchemeDefaultUserRole  sql.NullString
	TeamSchemeDefaultAdminRole sql.NullString
}

type teamMemberWithSchemeRolesList []teamMemberWithSchemeRoles

func teamMemberSliceColumns() []string {
	return []string{"TeamId", "UserId", "Roles", "DeleteAt", "SchemeUser", "SchemeAdmin", "SchemeGuest"}
}

func teamMemberToSlice(member *model.TeamMember) []interface{} {
	resultSlice := []interface{}{}
	resultSlice = append(resultSlice, member.TeamId)
	resultSlice = append(resultSlice, member.UserId)
	resultSlice = append(resultSlice, member.ExplicitRoles)
	resultSlice = append(resultSlice, member.DeleteAt)
	resultSlice = append(resultSlice, member.SchemeUser)
	resultSlice = append(resultSlice, member.SchemeAdmin)
	resultSlice = append(resultSlice, member.SchemeGuest)
	return resultSlice
}

func wildcardSearchTerm(term string) string {
	return strings.ToLower("%" + term + "%")
}

type rolesInfo struct {
	roles         []string
	explicitRoles []string
	schemeGuest   bool
	schemeUser    bool
	schemeAdmin   bool
}

func newSqlTeamStore(sqlStore SqlStore) store.TeamStore {
	s := &SqlTeamStore{
		SqlStore: sqlStore,
	}

	s.teamsQuery = s.getQueryBuilder().
		Select("Teams.*").
		From("Teams")

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Team{}, "Teams").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("DisplayName").SetMaxSize(64)
		table.ColMap("Name").SetMaxSize(64).SetUnique(true)
		table.ColMap("Description").SetMaxSize(255)
		table.ColMap("Type").SetMaxSize(255)
		table.ColMap("Email").SetMaxSize(128)
		table.ColMap("CompanyName").SetMaxSize(64)
		table.ColMap("AllowedDomains").SetMaxSize(1000)
		table.ColMap("InviteId").SetMaxSize(32)
		table.ColMap("SchemeId").SetMaxSize(26)

		tablem := db.AddTableWithName(teamMember{}, "TeamMembers").SetKeys(false, "TeamId", "UserId")
		tablem.ColMap("TeamId").SetMaxSize(26)
		tablem.ColMap("UserId").SetMaxSize(26)
		tablem.ColMap("Roles").SetMaxSize(64)
	}

	return s
}

// Save adds the team to the database if a team with the same name does not already
// exist in the database. It returns the team added if the operation is successful.
func (s SqlTeamStore) Save(team *model.Team) (*model.Team, error) {
	if len(team.Id) > 0 {
		return nil, store.NewErrInvalidInput("Team", "id", team.Id)
	}

	team.PreSave()

	if err := team.IsValid(); err != nil {
		return nil, err
	}

	if err := s.GetMaster().Insert(team); err != nil {
		if IsUniqueConstraintError(err, []string{"Name", "teams_name_key"}) {
			return nil, store.NewErrInvalidInput("Team", "id", team.Id)
		}
		return nil, errors.Wrapf(err, "failed to save Team with id=%s", team.Id)
	}
	return team, nil
}

func getTeamRoles(schemeGuest, schemeUser, schemeAdmin bool, defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole string, roles []string) rolesInfo {
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
		case model.TEAM_GUEST_ROLE_ID:
			result.schemeGuest = true
		case model.TEAM_USER_ROLE_ID:
			result.schemeUser = true
		case model.TEAM_ADMIN_ROLE_ID:
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
		if defaultTeamGuestRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamGuestRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.TEAM_GUEST_ROLE_ID)
		}
	}
	if result.schemeUser {
		if defaultTeamUserRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamUserRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.TEAM_USER_ROLE_ID)
		}
	}
	if result.schemeAdmin {
		if defaultTeamAdminRole != "" {
			schemeImpliedRoles = append(schemeImpliedRoles, defaultTeamAdminRole)
		} else {
			schemeImpliedRoles = append(schemeImpliedRoles, model.TEAM_ADMIN_ROLE_ID)
		}
	}
	for _, impliedRole := range schemeImpliedRoles {
		alreadyThere := false
		for _, role := range result.roles {
			if role == impliedRole {
				alreadyThere = true
			}
		}
		if !alreadyThere {
			result.roles = append(result.roles, impliedRole)
		}
	}
	return result
}

func (db teamMemberWithSchemeRoles) ToModel() *model.TeamMember {
	// Identify any scheme derived roles that are in "Roles" field due to not yet being migrated, and exclude
	// them from ExplicitRoles field.
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

	rolesResult := getTeamRoles(schemeGuest, schemeUser, schemeAdmin, defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole, strings.Fields(db.Roles))

	tm := &model.TeamMember{
		TeamId:        db.TeamId,
		UserId:        db.UserId,
		Roles:         strings.Join(rolesResult.roles, " "),
		DeleteAt:      db.DeleteAt,
		SchemeGuest:   rolesResult.schemeGuest,
		SchemeUser:    rolesResult.schemeUser,
		SchemeAdmin:   rolesResult.schemeAdmin,
		ExplicitRoles: strings.Join(rolesResult.explicitRoles, " "),
	}
	return tm
}

func (db teamMemberWithSchemeRolesList) ToModel() []*model.TeamMember {
	tms := make([]*model.TeamMember, 0)

	for _, tm := range db {
		tms = append(tms, tm.ToModel())
	}

	return tms
}

func (s SqlTeamStore) getTeamMembersWithSchemeSelectQuery() sq.SelectBuilder {
	return s.getQueryBuilder().
		Select(
			"TeamMembers.*",
			"TeamScheme.DefaultTeamGuestRole TeamSchemeDefaultGuestRole",
			"TeamScheme.DefaultTeamUserRole TeamSchemeDefaultUserRole",
			"TeamScheme.DefaultTeamAdminRole TeamSchemeDefaultAdminRole",
		).
		From("TeamMembers").
		LeftJoin("Teams ON TeamMembers.TeamId = Teams.Id").
		LeftJoin("Schemes TeamScheme ON Teams.SchemeId = TeamScheme.Id")
}

func (s SqlTeamStore) GetMember(teamId string, userId string) (*model.TeamMember, *model.AppError) {
	query := s.getTeamMembersWithSchemeSelectQuery().
		Where(sq.Eq{"TeamMembers.TeamId": teamId}).
		Where(sq.Eq{"TeamMembers.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlTeamStore.GetMember", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMember teamMemberWithSchemeRoles
	err = s.GetReplica().SelectOne(&dbMember, queryString, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, model.NewAppError("SqlTeamStore.GetMember", "store.sql_team.get_member.missing.app_error", nil, "teamId="+teamId+" userId="+userId+" "+err.Error(), http.StatusNotFound)
		}
		return nil, model.NewAppError("SqlTeamStore.GetMember", "store.sql_team.get_member.app_error", nil, "teamId="+teamId+" userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMember.ToModel(), nil
}

func (s SqlTeamStore) SaveMultipleMembers(members []*model.TeamMember, maxUsersPerTeam int) ([]*model.TeamMember, error) {
	newTeamMembers := map[string]int{}
	users := map[string]bool{}
	for _, member := range members {
		newTeamMembers[member.TeamId] = 0
	}

	for _, member := range members {
		newTeamMembers[member.TeamId]++
		users[member.UserId] = true

		if err := member.IsValid(); err != nil {
			return nil, err
		}
	}

	teams := []string{}
	for team := range newTeamMembers {
		teams = append(teams, team)
	}

	defaultTeamRolesByTeam := map[string]struct {
		Id    string
		Guest sql.NullString
		User  sql.NullString
		Admin sql.NullString
	}{}

	queryRoles := s.getQueryBuilder().
		Select(
			"Teams.Id as Id",
			"TeamScheme.DefaultTeamGuestRole as Guest",
			"TeamScheme.DefaultTeamUserRole as User",
			"TeamScheme.DefaultTeamAdminRole as Admin",
		).
		From("Teams").
		LeftJoin("Schemes TeamScheme ON Teams.SchemeId = TeamScheme.Id").
		Where(sq.Eq{"Teams.Id": teams})

	sqlRolesQuery, argsRoles, err := queryRoles.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "team_roles_tosql")
	}
	var defaultTeamsRoles []struct {
		Id    string
		Guest sql.NullString
		User  sql.NullString
		Admin sql.NullString
	}
	_, err = s.GetMaster().Select(&defaultTeamsRoles, sqlRolesQuery, argsRoles...)
	if err != nil {
		return nil, errors.Wrap(err, "default_team_roles_select")
	}

	for _, defaultRoles := range defaultTeamsRoles {
		defaultTeamRolesByTeam[defaultRoles.Id] = defaultRoles
	}

	if maxUsersPerTeam >= 0 {
		queryCount := s.getQueryBuilder().
			Select(
				"COUNT(0) as Count, TeamMembers.TeamId as TeamId",
			).
			From("TeamMembers").
			Join("Users ON TeamMembers.UserId = Users.Id").
			Where(sq.Eq{"TeamMembers.TeamId": teams}).
			Where(sq.Eq{"TeamMembers.DeleteAt": 0}).
			Where(sq.Eq{"Users.DeleteAt": 0}).
			GroupBy("TeamMembers.TeamId")

		sqlCountQuery, argsCount, errCount := queryCount.ToSql()
		if errCount != nil {
			return nil, errors.Wrap(err, "member_count_tosql")
		}

		var counters []struct {
			Count  int    `db:"Count"`
			TeamId string `db:"TeamId"`
		}

		_, err = s.GetMaster().Select(&counters, sqlCountQuery, argsCount...)
		if err != nil {
			return nil, errors.Wrap(err, "failed to count users in the teams of the memberships")
		}

		for teamId, newMembers := range newTeamMembers {
			existingMembers := 0
			for _, counter := range counters {
				if counter.TeamId == teamId {
					existingMembers = counter.Count
				}
			}
			if existingMembers+newMembers > maxUsersPerTeam {
				return nil, store.NewErrLimitExceeded("TeamMember", existingMembers+newMembers, "team members limit exceeded")
			}
		}
	}

	query := s.getQueryBuilder().Insert("TeamMembers").Columns(teamMemberSliceColumns()...)
	for _, member := range members {
		query = query.Values(teamMemberToSlice(member)...)
	}

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "insert_members_to_sql")
	}

	if _, err = s.GetMaster().Exec(sql, args...); err != nil {
		if IsUniqueConstraintError(err, []string{"TeamId", "teammembers_pkey", "PRIMARY"}) {
			return nil, store.NewErrConflict("TeamMember", err, "")
		}
		return nil, errors.Wrap(err, "unable_to_save_team_member")
	}

	newMembers := []*model.TeamMember{}
	for _, member := range members {
		s.InvalidateAllTeamIdsForUser(member.UserId)
		defaultTeamGuestRole := defaultTeamRolesByTeam[member.TeamId].Guest.String
		defaultTeamUserRole := defaultTeamRolesByTeam[member.TeamId].User.String
		defaultTeamAdminRole := defaultTeamRolesByTeam[member.TeamId].Admin.String
		rolesResult := getTeamRoles(member.SchemeGuest, member.SchemeUser, member.SchemeAdmin, defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole, strings.Fields(member.ExplicitRoles))
		newMember := *member
		newMember.SchemeGuest = rolesResult.schemeGuest
		newMember.SchemeUser = rolesResult.schemeUser
		newMember.SchemeAdmin = rolesResult.schemeAdmin
		newMember.Roles = strings.Join(rolesResult.roles, " ")
		newMember.ExplicitRoles = strings.Join(rolesResult.explicitRoles, " ")
		newMembers = append(newMembers, &newMember)
	}

	return newMembers, nil
}

func applyTeamMemberViewRestrictionsFilterForStats(query sq.SelectBuilder, teamId string, restrictions *model.ViewUsersRestrictions) sq.SelectBuilder {
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
		resultQuery = resultQuery.Join(fmt.Sprintf("TeamMembers rtm ON ( rtm.UserId = Users.Id AND rtm.DeleteAt = 0 AND rtm.TeamId IN (%s))", sq.Placeholders(len(teams))), teams...)
	}
	if restrictions.Channels != nil && len(restrictions.Channels) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("ChannelMembers rcm ON ( rcm.UserId = Users.Id AND rcm.ChannelId IN (%s))", sq.Placeholders(len(channels))), channels...)
	}

	return resultQuery
}

func applyTeamMemberViewRestrictionsFilter(query sq.SelectBuilder, teamId string, restrictions *model.ViewUsersRestrictions) sq.SelectBuilder {
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

	resultQuery := query.Join("Users ru ON (TeamMembers.UserId = ru.Id)")
	if restrictions.Teams != nil && len(restrictions.Teams) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("TeamMembers rtm ON ( rtm.UserId = ru.Id AND rtm.DeleteAt = 0 AND rtm.TeamId IN (%s))", sq.Placeholders(len(teams))), teams...)
	}
	if restrictions.Channels != nil && len(restrictions.Channels) > 0 {
		resultQuery = resultQuery.Join(fmt.Sprintf("ChannelMembers rcm ON ( rcm.UserId = ru.Id AND rcm.ChannelId IN (%s))", sq.Placeholders(len(channels))), channels...)
	}

	return resultQuery.Distinct()
}

func (s SqlTeamStore) GetMembers(teamId string, offset int, limit int, teamMembersGetOptions *model.TeamMembersGetOptions) ([]*model.TeamMember, *model.AppError) {
	query := s.getTeamMembersWithSchemeSelectQuery().
		Where(sq.Eq{"TeamMembers.TeamId": teamId}).
		Where(sq.Eq{"TeamMembers.DeleteAt": 0}).
		Limit(uint64(limit)).
		Offset(uint64(offset))

	if teamMembersGetOptions == nil || teamMembersGetOptions.Sort == "" {
		query = query.OrderBy("UserId")
	}

	if teamMembersGetOptions != nil {
		if teamMembersGetOptions.Sort == model.USERNAME || teamMembersGetOptions.ExcludeDeletedUsers {
			query = query.LeftJoin("Users ON TeamMembers.UserId = Users.Id")
		}

		if teamMembersGetOptions.ExcludeDeletedUsers {
			query = query.Where(sq.Eq{"Users.DeleteAt": 0})
		}

		if teamMembersGetOptions.Sort == model.USERNAME {
			query = query.OrderBy(model.USERNAME)
		}

		query = applyTeamMemberViewRestrictionsFilter(query, teamId, teamMembersGetOptions.ViewRestrictions)
	}

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlTeamStore.GetMembers", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers teamMemberWithSchemeRolesList
	_, err = s.GetReplica().Select(&dbMembers, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlTeamStore.GetMembers", "store.sql_team.get_members.app_error", nil, "teamId="+teamId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMembers.ToModel(), nil
}

func (s SqlTeamStore) SaveMember(member *model.TeamMember, maxUsersPerTeam int) (*model.TeamMember, error) {
	members, err := s.SaveMultipleMembers([]*model.TeamMember{member}, maxUsersPerTeam)
	if err != nil {
		return nil, err
	}
	return members[0], nil
}

func (s SqlTeamStore) InvalidateAllTeamIdsForUser(userId string) {}

// GetUserTeamIds get the team ids to which the user belongs to. allowFromCache parameter does not have any effect in this Store
func (s SqlTeamStore) GetUserTeamIds(userID string, allowFromCache bool) ([]string, *model.AppError) {
	var teamIds []string
	query, args, err := s.getQueryBuilder().
		Select("TeamId").
		From("TeamMembers").
		Join("Teams ON TeamMembers.TeamId = Teams.Id").
		Where(sq.Eq{"TeamMembers.UserId": userID, "TeamMembers.DeleteAt": 0, "Teams.DeleteAt": 0}).ToSql()

	if err != nil {
		return []string{}, model.NewAppError("SqlTeamStore.GetUserTeamIds", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	_, err = s.GetReplica().Select(&teamIds, query, args...)
	if err != nil {
		return []string{}, model.NewAppError("SqlTeamStore.GetUserTeamIds", "store.sql_team.get_user_team_ids.app_error", nil, "userID="+userID+" "+err.Error(), http.StatusInternalServerError)
	}

	return teamIds, nil
}

func (s SqlTeamStore) GetTeamsForUser(userId string) ([]*model.TeamMember, *model.AppError) {
	query := s.getTeamMembersWithSchemeSelectQuery().
		Where(sq.Eq{"TeamMembers.UserId": userId})

	queryString, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlTeamStore.GetMembers", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	var dbMembers teamMemberWithSchemeRolesList
	_, err = s.GetReplica().Select(&dbMembers, queryString, args...)
	if err != nil {
		return nil, model.NewAppError("SqlTeamStore.GetMembers", "store.sql_team.get_members.app_error", nil, "userId="+userId+" "+err.Error(), http.StatusInternalServerError)
	}

	return dbMembers.ToModel(), nil
}

func (s SqlTeamStore) GetActiveMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError) {
	query := s.getQueryBuilder().
		Select("count(DISTINCT TeamMembers.UserId)").
		From("TeamMembers, Users").
		Where("TeamMembers.DeleteAt = 0").
		Where("TeamMembers.UserId = Users.Id").
		Where("Users.DeleteAt = 0").
		Where(sq.Eq{"TeamMembers.TeamId": teamId})

	query = applyTeamMemberViewRestrictionsFilterForStats(query, teamId, restrictions)
	queryString, args, err := query.ToSql()
	if err != nil {
		return 0, model.NewAppError("SqlTeamStore.GetActiveMemberCount", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	count, err := s.GetReplica().SelectInt(queryString, args...)
	if err != nil {
		return 0, model.NewAppError("SqlTeamStore.GetActiveMemberCount", "store.sql_team.get_active_member_count.app_error", nil, "teamId="+teamId+" "+err.Error(), http.StatusInternalServerError)
	}

	return count, nil
}

func (s SqlTeamStore) UpdateMember(member *model.TeamMember) (*model.TeamMember, *model.AppError) {
	members, err := s.UpdateMultipleMembers([]*model.TeamMember{member})
	if err != nil {
		return nil, err
	}
	return members[0], nil
}

func (s SqlTeamStore) UpdateMultipleMembers(members []*model.TeamMember) ([]*model.TeamMember, *model.AppError) {
	teams := []string{}
	for _, member := range members {
		member.PreUpdate()

		if err := member.IsValid(); err != nil {
			return nil, err
		}

		if _, err := s.GetMaster().Update(NewTeamMemberFromModel(member)); err != nil {
			return nil, model.NewAppError("SqlTeamStore.UpdateMember", "store.sql_team.save_member.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
		teams = append(teams, member.TeamId)
	}

	query := s.getQueryBuilder().
		Select(
			"Teams.Id as Id",
			"TeamScheme.DefaultTeamGuestRole as Guest",
			"TeamScheme.DefaultTeamUserRole as User",
			"TeamScheme.DefaultTeamAdminRole as Admin",
		).
		From("Teams").
		LeftJoin("Schemes TeamScheme ON Teams.SchemeId = TeamScheme.Id").
		Where(sq.Eq{"Teams.Id": teams})

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	var defaultTeamsRoles []struct {
		Id    string
		Guest sql.NullString
		User  sql.NullString
		Admin sql.NullString
	}
	_, err = s.GetMaster().Select(&defaultTeamsRoles, sqlQuery, args...)
	if err != nil {
		return nil, model.NewAppError("SqlUserStore.Save", "store.sql_user.save.member_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	defaultTeamRolesByTeam := map[string]struct {
		Id    string
		Guest sql.NullString
		User  sql.NullString
		Admin sql.NullString
	}{}
	for _, defaultRoles := range defaultTeamsRoles {
		defaultTeamRolesByTeam[defaultRoles.Id] = defaultRoles
	}

	updatedMembers := []*model.TeamMember{}
	for _, member := range members {
		s.InvalidateAllTeamIdsForUser(member.UserId)
		defaultTeamGuestRole := defaultTeamRolesByTeam[member.TeamId].Guest.String
		defaultTeamUserRole := defaultTeamRolesByTeam[member.TeamId].User.String
		defaultTeamAdminRole := defaultTeamRolesByTeam[member.TeamId].Admin.String
		rolesResult := getTeamRoles(member.SchemeGuest, member.SchemeUser, member.SchemeAdmin, defaultTeamGuestRole, defaultTeamUserRole, defaultTeamAdminRole, strings.Fields(member.ExplicitRoles))
		updatedMember := *member
		updatedMember.SchemeGuest = rolesResult.schemeGuest
		updatedMember.SchemeUser = rolesResult.schemeUser
		updatedMember.SchemeAdmin = rolesResult.schemeAdmin
		updatedMember.Roles = strings.Join(rolesResult.roles, " ")
		updatedMember.ExplicitRoles = strings.Join(rolesResult.explicitRoles, " ")
		updatedMembers = append(updatedMembers, &updatedMember)
	}

	return updatedMembers, nil
}

func (s SqlTeamStore) UserBelongsToTeams(userId string, teamIds []string) (bool, *model.AppError) {
	idQuery := sq.Eq{
		"UserId":   userId,
		"TeamId":   teamIds,
		"DeleteAt": 0,
	}

	query, params, err := s.getQueryBuilder().Select("Count(*)").From("TeamMembers").Where(idQuery).ToSql()
	if err != nil {
		return false, model.NewAppError("SqlTeamStore.UserBelongsToTeams", "store.sql.build_query.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	c, err := s.GetReplica().SelectInt(query, params...)
	if err != nil {
		return false, model.NewAppError("SqlTeamStore.UserBelongsToTeams", "store.sql_team.user_belongs_to_teams.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return c > 0, nil
}
