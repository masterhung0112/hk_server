package store

import (
	"github.com/masterhung0112/go_server/model"
)

type Store interface {
  Team() TeamStore
  User() UserStore
  System() SystemStore
  Role() RoleStore
  Scheme() SchemeStore
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
  InferSystemInstallDate() (int64, *model.AppError)
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

type RoleStore interface {
	Save(role *model.Role) (*model.Role, error)
	Get(roleId string) (*model.Role, error)
	GetAll() ([]*model.Role, error)
	GetByName(name string) (*model.Role, error)
	GetByNames(names []string) ([]*model.Role, error)
	Delete(roleId string) (*model.Role, error)
	PermanentDeleteAll() error

	// HigherScopedPermissions retrieves the higher-scoped permissions of a list of role names. The higher-scope
	// (either team scheme or system scheme) is determined based on whether the team has a scheme or not.
	ChannelHigherScopedPermissions(roleNames []string) (map[string]*model.RolePermissions, error)

	// AllChannelSchemeRoles returns all of the roles associated to channel schemes.
	AllChannelSchemeRoles() ([]*model.Role, error)

	// ChannelRolesUnderTeamRole returns all of the non-deleted roles that are affected by updates to the
	// given role.
	ChannelRolesUnderTeamRole(roleName string) ([]*model.Role, error)
}

type TeamStore interface {
	// Save(team *model.Team) (*model.Team, error)
	// Update(team *model.Team) (*model.Team, error)
	// Get(id string) (*model.Team, error)
	// GetByName(name string) (*model.Team, error)
	// GetByNames(name []string) ([]*model.Team, error)
	// SearchAll(term string, opts *model.TeamSearch) ([]*model.Team, error)
	// SearchAllPaged(term string, opts *model.TeamSearch) ([]*model.Team, int64, error)
	// SearchOpen(term string) ([]*model.Team, error)
	// SearchPrivate(term string) ([]*model.Team, error)
	// GetAll() ([]*model.Team, error)
	// GetAllPage(offset int, limit int) ([]*model.Team, error)
	// GetAllPrivateTeamListing() ([]*model.Team, error)
	// GetAllPrivateTeamPageListing(offset int, limit int) ([]*model.Team, error)
	// GetAllPublicTeamPageListing(offset int, limit int) ([]*model.Team, error)
	// GetAllTeamListing() ([]*model.Team, error)
	// GetAllTeamPageListing(offset int, limit int) ([]*model.Team, error)
	// GetTeamsByUserId(userId string) ([]*model.Team, error)
	// GetByInviteId(inviteId string) (*model.Team, error)
	// PermanentDelete(teamId string) error
	// AnalyticsTeamCount(includeDeleted bool) (int64, error)
	// AnalyticsPublicTeamCount() (int64, error)
	// AnalyticsPrivateTeamCount() (int64, error)
	SaveMultipleMembers(members []*model.TeamMember, maxUsersPerTeam int) ([]*model.TeamMember, error)
	SaveMember(member *model.TeamMember, maxUsersPerTeam int) (*model.TeamMember, error)
	// UpdateMember(member *model.TeamMember) (*model.TeamMember, *model.AppError)
	// UpdateMultipleMembers(members []*model.TeamMember) ([]*model.TeamMember, *model.AppError)
	// GetMember(teamId string, userId string) (*model.TeamMember, *model.AppError)
	GetMembers(teamId string, offset int, limit int, teamMembersGetOptions *model.TeamMembersGetOptions) ([]*model.TeamMember, *model.AppError)
	// GetMembersByIds(teamId string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.TeamMember, *model.AppError)
	// GetTotalMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError)
	// GetActiveMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError)
	// GetTeamsForUser(userId string) ([]*model.TeamMember, *model.AppError)
	// GetTeamsForUserWithPagination(userId string, page, perPage int) ([]*model.TeamMember, *model.AppError)
	// GetChannelUnreadsForAllTeams(excludeTeamId, userId string) ([]*model.ChannelUnread, *model.AppError)
	// GetChannelUnreadsForTeam(teamId, userId string) ([]*model.ChannelUnread, *model.AppError)
	// RemoveMember(teamId string, userId string) *model.AppError
	// RemoveMembers(teamId string, userIds []string) *model.AppError
	// RemoveAllMembersByTeam(teamId string) *model.AppError
	// RemoveAllMembersByUser(userId string) *model.AppError
	// UpdateLastTeamIconUpdate(teamId string, curTime int64) *model.AppError
	// GetTeamsByScheme(schemeId string, offset int, limit int) ([]*model.Team, *model.AppError)
	// MigrateTeamMembers(fromTeamId string, fromUserId string) (map[string]string, *model.AppError)
	// ResetAllTeamSchemes() *model.AppError
	// ClearAllCustomRoleAssignments() *model.AppError
	// AnalyticsGetTeamCountForScheme(schemeId string) (int64, *model.AppError)
	// GetAllForExportAfter(limit int, afterId string) ([]*model.TeamForExport, *model.AppError)
	// GetTeamMembersForExport(userId string) ([]*model.TeamMemberForExport, *model.AppError)
	// UserBelongsToTeams(userId string, teamIds []string) (bool, *model.AppError)
	// GetUserTeamIds(userId string, allowFromCache bool) ([]string, *model.AppError)
	InvalidateAllTeamIdsForUser(userId string)
	// ClearCaches()

	// // UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// // non-admin members.
	// UpdateMembersRole(teamID string, userIDs []string) *model.AppError

	// // GroupSyncedTeamCount returns the count of non-deleted group-constrained teams.
	// GroupSyncedTeamCount() (int64, *model.AppError)
}

type SchemeStore interface {
	Save(scheme *model.Scheme) (*model.Scheme, error)
	// Get(schemeId string) (*model.Scheme, error)
	// GetByName(schemeName string) (*model.Scheme, error)
	// GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error)
	// Delete(schemeId string) (*model.Scheme, error)
	// PermanentDeleteAll() error
	// CountByScope(scope string) (int64, error)
	// CountWithoutPermission(scope, permissionID string, roleScope model.RoleScope, roleType model.RoleType) (int64, error)
}