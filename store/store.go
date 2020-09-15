package store

import (
	"github.com/masterhung0112/go_server/model"
)

type StoreResult struct {
	Data interface{}
	Err  *model.AppError

	// NErr a temporary field used by the new code for the AppError migration. This will later become Err when the entire store is migrated.
	NErr error
}

type Store interface {
	Team() TeamStore
	Channel() ChannelStore
	User() UserStore
	System() SystemStore
	Role() RoleStore
	Scheme() SchemeStore
	Session() SessionStore
	UserAccessToken() UserAccessTokenStore
	Token() TokenStore
	Close()
	DropAllTables()
	MarkSystemRanUnitTests()
}

type UserStore interface {
	Save(user *model.User) (*model.User, *model.AppError)
	Update(user *model.User, allowRoleUpdate bool) (*model.UserUpdate, *model.AppError)
	Get(id string) (*model.User, *model.AppError)
	GetAll() ([]*model.User, *model.AppError)
	Count(options model.UserCountOptions) (int64, *model.AppError)
	PermanentDelete(userId string) *model.AppError
	InferSystemInstallDate() (int64, *model.AppError)

	GetByUsername(username string) (*model.User, *model.AppError)
	GetByEmail(email string) (*model.User, *model.AppError)
	GetProfilesWithoutTeam(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfilesNotInChannel(teamId string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetEtagForProfilesNotInTeam(teamId string) string
	GetProfilesNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError)
	GetEtagForProfiles(teamId string) string
	GetProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	GetProfilesInChannel(channelId string, offset int, limit int) ([]*model.User, *model.AppError)
	GetProfilesInChannelByStatus(channelId string, offset int, limit int) ([]*model.User, *model.AppError)
	GetAllProfiles(options *model.UserGetOptions) ([]*model.User, *model.AppError)
	UpdateFailedPasswordAttempts(userId string, attempts int) *model.AppError
	GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, *model.AppError)
	GetProfileByIds(userIds []string, options *UserGetByIdsOpts, allowFromCache bool) ([]*model.User, *model.AppError)
	GetChannelGroupUsers(channelID string) ([]*model.User, *model.AppError)
	UpdateUpdateAt(userId string) (int64, *model.AppError)
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
	UpdateMember(member *model.TeamMember) (*model.TeamMember, *model.AppError)
	UpdateMultipleMembers(members []*model.TeamMember) ([]*model.TeamMember, *model.AppError)
	GetMember(teamId string, userId string) (*model.TeamMember, *model.AppError)
	GetMembers(teamId string, offset int, limit int, teamMembersGetOptions *model.TeamMembersGetOptions) ([]*model.TeamMember, *model.AppError)
	// GetMembersByIds(teamId string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.TeamMember, *model.AppError)
	// GetTotalMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError)
	GetActiveMemberCount(teamId string, restrictions *model.ViewUsersRestrictions) (int64, *model.AppError)
	GetTeamsForUser(userId string) ([]*model.TeamMember, *model.AppError)
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
	GetUserTeamIds(userId string, allowFromCache bool) ([]string, *model.AppError)
	InvalidateAllTeamIdsForUser(userId string)
	// ClearCaches()

	// // UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// // non-admin members.
	// UpdateMembersRole(teamID string, userIDs []string) *model.AppError

	// // GroupSyncedTeamCount returns the count of non-deleted group-constrained teams.
	// GroupSyncedTeamCount() (int64, *model.AppError)
}

type ChannelStore interface {
	Save(channel *model.Channel, maxChannelsPerTeam int64) (*model.Channel, error)
	// CreateDirectChannel(userId *model.User, otherUserId *model.User) (*model.Channel, error)
	// SaveDirectChannel(channel *model.Channel, member1 *model.ChannelMember, member2 *model.ChannelMember) (*model.Channel, error)
	// Update(channel *model.Channel) (*model.Channel, error)
	// UpdateSidebarChannelCategoryOnMove(channel *model.Channel, newTeamId string) error
	// ClearSidebarOnTeamLeave(userId, teamId string) error
	Get(id string, allowFromCache bool) (*model.Channel, error)
	// InvalidateChannel(id string)
	// InvalidateChannelByName(teamId, name string)
	// GetFromMaster(id string) (*model.Channel, error)
	// Delete(channelId string, time int64) error
	// Restore(channelId string, time int64) error
	// SetDeleteAt(channelId string, deleteAt int64, updateAt int64) error
	// PermanentDelete(channelId string) error
	// PermanentDeleteByTeam(teamId string) error
	GetByName(team_id string, name string, allowFromCache bool) (*model.Channel, error)
	// GetByNames(team_id string, names []string, allowFromCache bool) ([]*model.Channel, error)
	// GetByNameIncludeDeleted(team_id string, name string, allowFromCache bool) (*model.Channel, error)
	// GetDeletedByName(team_id string, name string) (*model.Channel, error)
	// GetDeleted(team_id string, offset int, limit int, userId string) (*model.ChannelList, error)
	// GetChannels(teamId string, userId string, includeDeleted bool, lastDeleteAt int) (*model.ChannelList, error)
	// GetAllChannels(page, perPage int, opts ChannelSearchOpts) (*model.ChannelListWithTeamData, error)
	// GetAllChannelsCount(opts ChannelSearchOpts) (int64, error)
	// GetMoreChannels(teamId string, userId string, offset int, limit int) (*model.ChannelList, error)
	// GetPrivateChannelsForTeam(teamId string, offset int, limit int) (*model.ChannelList, error)
	// GetPublicChannelsForTeam(teamId string, offset int, limit int) (*model.ChannelList, error)
	// GetPublicChannelsByIdsForTeam(teamId string, channelIds []string) (*model.ChannelList, error)
	// GetChannelCounts(teamId string, userId string) (*model.ChannelCounts, error)
	// GetTeamChannels(teamId string) (*model.ChannelList, error)
	// GetAll(teamId string) ([]*model.Channel, error)
	// GetChannelsByIds(channelIds []string, includeDeleted bool) ([]*model.Channel, error)
	GetForPost(postId string) (*model.Channel, error)
	SaveMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, *model.AppError)
	SaveMember(member *model.ChannelMember) (*model.ChannelMember, *model.AppError)
	// UpdateMember(member *model.ChannelMember) (*model.ChannelMember, *model.AppError)
	// UpdateMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, *model.AppError)
	// GetMembers(channelId string, offset, limit int) (*model.ChannelMembers, *model.AppError)
	GetMember(channelId string, userId string) (*model.ChannelMember, *model.AppError)
	// GetChannelMembersTimezones(channelId string) ([]model.StringMap, *model.AppError)
	GetAllChannelMembersForUser(userId string, allowFromCache bool, includeDeleted bool) (map[string]string, *model.AppError)
	InvalidateAllChannelMembersForUser(userId string)
	// IsUserInChannelUseCache(userId string, channelId string) bool
	// GetAllChannelMembersNotifyPropsForChannel(channelId string, allowFromCache bool) (map[string]model.StringMap, *model.AppError)
	// InvalidateCacheForChannelMembersNotifyProps(channelId string)
	GetMemberForPost(postId string, userId string) (*model.ChannelMember, *model.AppError)
	// InvalidateMemberCount(channelId string)
	// GetMemberCountFromCache(channelId string) int64
	// GetMemberCount(channelId string, allowFromCache bool) (int64, *model.AppError)
	// GetMemberCountsByGroup(channelID string, includeTimezones bool) ([]*model.ChannelMemberCountByGroup, *model.AppError)
	// InvalidatePinnedPostCount(channelId string)
	// GetPinnedPostCount(channelId string, allowFromCache bool) (int64, *model.AppError)
	// InvalidateGuestCount(channelId string)
	// GetGuestCount(channelId string, allowFromCache bool) (int64, *model.AppError)
	// GetPinnedPosts(channelId string) (*model.PostList, *model.AppError)
	// RemoveMember(channelId string, userId string) *model.AppError
	// RemoveMembers(channelId string, userIds []string) *model.AppError
	// PermanentDeleteMembersByUser(userId string) *model.AppError
	// PermanentDeleteMembersByChannel(channelId string) *model.AppError
	// UpdateLastViewedAt(channelIds []string, userId string) (map[string]int64, *model.AppError)
	// UpdateLastViewedAtPost(unreadPost *model.Post, userID string, mentionCount int) (*model.ChannelUnreadAt, *model.AppError)
	// CountPostsAfter(channelId string, timestamp int64, userId string) (int, *model.AppError)
	// IncrementMentionCount(channelId string, userId string) *model.AppError
	// AnalyticsTypeCount(teamId string, channelType string) (int64, *model.AppError)
	// GetMembersForUser(teamId string, userId string) (*model.ChannelMembers, *model.AppError)
	// GetMembersForUserWithPagination(teamId, userId string, page, perPage int) (*model.ChannelMembers, *model.AppError)
	// AutocompleteInTeam(teamId string, term string, includeDeleted bool) (*model.ChannelList, *model.AppError)
	// AutocompleteInTeamForSearch(teamId string, userId string, term string, includeDeleted bool) (*model.ChannelList, *model.AppError)
	// SearchAllChannels(term string, opts ChannelSearchOpts) (*model.ChannelListWithTeamData, int64, *model.AppError)
	// SearchInTeam(teamId string, term string, includeDeleted bool) (*model.ChannelList, *model.AppError)
	// SearchArchivedInTeam(teamId string, term string, userId string) (*model.ChannelList, *model.AppError)
	// SearchForUserInTeam(userId string, teamId string, term string, includeDeleted bool) (*model.ChannelList, *model.AppError)
	// SearchMore(userId string, teamId string, term string) (*model.ChannelList, *model.AppError)
	// SearchGroupChannels(userId, term string) (*model.ChannelList, *model.AppError)
	// GetMembersByIds(channelId string, userIds []string) (*model.ChannelMembers, *model.AppError)
	// AnalyticsDeletedTypeCount(teamId string, channelType string) (int64, *model.AppError)
	// GetChannelUnread(channelId, userId string) (*model.ChannelUnread, *model.AppError)
	// ClearCaches()
	// GetChannelsByScheme(schemeId string, offset int, limit int) (model.ChannelList, *model.AppError)
	// MigrateChannelMembers(fromChannelId string, fromUserId string) (map[string]string, *model.AppError)
	// ResetAllChannelSchemes() *model.AppError
	// ClearAllCustomRoleAssignments() *model.AppError
	// MigratePublicChannels() error
	CreateInitialSidebarCategories(userId, teamId string) error
	// GetSidebarCategories(userId, teamId string) (*model.OrderedSidebarCategories, *model.AppError)
	GetSidebarCategory(categoryId string) (*model.SidebarCategoryWithChannels, *model.AppError)
	// GetSidebarCategoryOrder(userId, teamId string) ([]string, *model.AppError)
	// CreateSidebarCategory(userId, teamId string, newCategory *model.SidebarCategoryWithChannels) (*model.SidebarCategoryWithChannels, *model.AppError)
	// UpdateSidebarCategoryOrder(userId, teamId string, categoryOrder []string) *model.AppError
	// UpdateSidebarCategories(userId, teamId string, categories []*model.SidebarCategoryWithChannels) ([]*model.SidebarCategoryWithChannels, *model.AppError)
	// UpdateSidebarChannelsByPreferences(preferences *model.Preferences) error
	// DeleteSidebarChannelsByPreferences(preferences *model.Preferences) error
	// DeleteSidebarCategory(categoryId string) *model.AppError
	// GetAllChannelsForExportAfter(limit int, afterId string) ([]*model.ChannelForExport, *model.AppError)
	// GetAllDirectChannelsForExportAfter(limit int, afterId string) ([]*model.DirectChannelForExport, *model.AppError)
	// GetChannelMembersForExport(userId string, teamId string) ([]*model.ChannelMemberForExport, *model.AppError)
	// RemoveAllDeactivatedMembers(channelId string) *model.AppError
	// GetChannelsBatchForIndexing(startTime, endTime int64, limit int) ([]*model.Channel, *model.AppError)
	// UserBelongsToChannels(userId string, channelIds []string) (bool, *model.AppError)

	// // UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// // non-admin members.
	// UpdateMembersRole(channelID string, userIDs []string) *model.AppError

	// // GroupSyncedChannelCount returns the count of non-deleted group-constrained channels.
	// GroupSyncedChannelCount() (int64, *model.AppError)
}

type SchemeStore interface {
	Save(scheme *model.Scheme) (*model.Scheme, error)
	// Get(schemeId string) (*model.Scheme, error)
	// GetByName(schemeName string) (*model.Scheme, error)
	GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error)
	// Delete(schemeId string) (*model.Scheme, error)
	// PermanentDeleteAll() error
	// CountByScope(scope string) (int64, error)
	// CountWithoutPermission(scope, permissionID string, roleScope model.RoleScope, roleType model.RoleType) (int64, error)
}

type SessionStore interface {
	Get(sessionIdOrToken string) (*model.Session, error)
	Save(session *model.Session) (*model.Session, error)
	GetSessions(userId string) ([]*model.Session, error)
	// GetSessionsWithActiveDeviceIds(userId string) ([]*model.Session, error)
	// GetSessionsExpired(thresholdMillis int64, mobileOnly bool, unnotifiedOnly bool) ([]*model.Session, error)
	// UpdateExpiredNotify(sessionid string, notified bool) error
	Remove(sessionIdOrToken string) error
	// RemoveAllSessions() error
	// PermanentDeleteSessionsByUser(teamId string) error
	// UpdateExpiresAt(sessionId string, time int64) error
	// UpdateLastActivityAt(sessionId string, time int64) error
	UpdateRoles(userId string, roles string) (string, error)
	// UpdateDeviceId(id string, deviceId string, expiresAt int64) (string, error)
	// UpdateProps(session *model.Session) error
	// AnalyticsSessionCount() (int64, error)
	// Cleanup(expiryTime int64, batchSize int64)
}

type UserAccessTokenStore interface {
	// Save(token *model.UserAccessToken) (*model.UserAccessToken, error)
	// DeleteAllForUser(userId string) error
	// Delete(tokenId string) error
	// Get(tokenId string) (*model.UserAccessToken, error)
	// GetAll(offset int, limit int) ([]*model.UserAccessToken, error)
	GetByToken(tokenString string) (*model.UserAccessToken, error)
	// GetByUser(userId string, page, perPage int) ([]*model.UserAccessToken, error)
	// Search(term string) ([]*model.UserAccessToken, error)
	// UpdateTokenEnable(tokenId string) error
	// UpdateTokenDisable(tokenId string) error
}

type TokenStore interface {
	Save(recovery *model.Token) error
	Delete(token string) error
	GetByToken(token string) (*model.Token, error)
	Cleanup()
	RemoveAllTokensByType(tokenType string) error
}

type UserGetByIdsOpts struct {
	// IsAdmin tracks whether or not the request is being made by an administrator. Does nothing when provided by a client.
	IsAdmin bool

	// Restrict to search in a list of teams and channels. Does nothing when provided by a client.
	ViewRestrictions *model.ViewUsersRestrictions

	// Since filters the users based on their UpdateAt timestamp.
	Since int64
}
