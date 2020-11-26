package store

import (
	"github.com/masterhung0112/hk_server/model"
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
	Post() PostStore
	Thread() ThreadStore
	User() UserStore
	System() SystemStore
	Role() RoleStore
	Scheme() SchemeStore
	Session() SessionStore
	UserAccessToken() UserAccessTokenStore
	Preference() PreferenceStore
	Token() TokenStore
	Group() GroupStore
	Close()
	DropAllTables()
	MarkSystemRanUnitTests()
	LockToMaster()
	UnlockFromMaster()
}

type ThreadStore interface {
	// SaveMultiple(thread []*model.Thread) ([]*model.Thread, int, error)
	// Save(thread *model.Thread) (*model.Thread, error)
	// Update(thread *model.Thread) (*model.Thread, error)
	// Get(id string) (*model.Thread, error)
	// GetThreadsForUser(userId string, opts model.GetUserThreadsOpts) (*model.Threads, error)
	// Delete(postId string) error

	// MarkAllAsRead(userId string, timestamp int64) error
	// MarkAsRead(userId, threadId string, timestamp int64) error

	// SaveMembership(membership *model.ThreadMembership) (*model.ThreadMembership, error)
	// UpdateMembership(membership *model.ThreadMembership) (*model.ThreadMembership, error)
	// GetMembershipsForUser(userId string) ([]*model.ThreadMembership, error)
	// GetMembershipForUser(userId, postId string) (*model.ThreadMembership, error)
	// DeleteMembershipForUser(userId, postId string) error
	// CreateMembershipIfNeeded(userId, postId string, following bool) error
	// CollectThreadsWithNewerReplies(userId string, channelIds []string, timestamp int64) ([]string, error)
	// UpdateUnreadsByChannel(userId string, changedThreads []string, timestamp int64) error
}

type PostStore interface {
	SaveMultiple(posts []*model.Post) ([]*model.Post, int, error)
	Save(post *model.Post) (*model.Post, error)
	// Update(newPost *model.Post, oldPost *model.Post) (*model.Post, error)
	Get(id string, skipFetchThreads bool) (*model.PostList, error)
	GetSingle(id string) (*model.Post, error)
	// Delete(postId string, time int64, deleteByID string) error
	// PermanentDeleteByUser(userId string) error
	// PermanentDeleteByChannel(channelId string) error
	// GetPosts(options model.GetPostsOptions, allowFromCache bool) (*model.PostList, error)
	// GetFlaggedPosts(userId string, offset int, limit int) (*model.PostList, error)
	// // @openTracingParams userId, teamId, offset, limit
	// GetFlaggedPostsForTeam(userId, teamId string, offset int, limit int) (*model.PostList, error)
	// GetFlaggedPostsForChannel(userId, channelId string, offset int, limit int) (*model.PostList, error)
	// GetPostsBefore(options model.GetPostsOptions) (*model.PostList, error)
	// GetPostsAfter(options model.GetPostsOptions) (*model.PostList, error)
	// GetPostsSince(options model.GetPostsSinceOptions, allowFromCache bool) (*model.PostList, error)
	// GetPostAfterTime(channelId string, time int64) (*model.Post, error)
	// GetPostIdAfterTime(channelId string, time int64) (string, error)
	// GetPostIdBeforeTime(channelId string, time int64) (string, error)
	// GetEtag(channelId string, allowFromCache bool) string
	// Search(teamId string, userId string, params *model.SearchParams) (*model.PostList, error)
	// AnalyticsUserCountsWithPostsByDay(teamId string) (model.AnalyticsRows, error)
	// AnalyticsPostCountsByDay(options *model.AnalyticsPostCountsOptions) (model.AnalyticsRows, error)
	// AnalyticsPostCount(teamId string, mustHaveFile bool, mustHaveHashtag bool) (int64, error)
	// ClearCaches()
	// InvalidateLastPostTimeCache(channelId string)
	// GetPostsCreatedAt(channelId string, time int64) ([]*model.Post, error)
	// Overwrite(post *model.Post) (*model.Post, error)
	// OverwriteMultiple(posts []*model.Post) ([]*model.Post, int, error)
	// GetPostsByIds(postIds []string) ([]*model.Post, error)
	// GetPostsBatchForIndexing(startTime int64, endTime int64, limit int) ([]*model.PostForIndexing, error)
	// PermanentDeleteBatch(endTime int64, limit int64) (int64, error)
	// GetOldest() (*model.Post, error)
	GetMaxPostSize() int
	// GetParentsForExportAfter(limit int, afterId string) ([]*model.PostForExport, error)
	// GetRepliesForExport(parentId string) ([]*model.ReplyForExport, error)
	// GetDirectPostParentsForExportAfter(limit int, afterId string) ([]*model.DirectPostForExport, error)
	// SearchPostsInTeamForUser(paramsList []*model.SearchParams, userId, teamId string, page, perPage int) (*model.PostSearchResults, error)
	// GetOldestEntityCreationTime() (int64, error)
}

type UserStore interface {
	Save(user *model.User) (*model.User, error)
	Update(user *model.User, allowRoleUpdate bool) (*model.UserUpdate, error)
	UpdateLastPictureUpdate(userId string) error
	ResetLastPictureUpdate(userId string) error
	UpdatePassword(userId, newPassword string) error
	UpdateUpdateAt(userId string) (int64, error)
	UpdateAuthData(userId string, service string, authData *string, email string, resetMfa bool) (string, error)
	UpdateMfaSecret(userId, secret string) error
	UpdateMfaActive(userId string, active bool) error
	Get(id string) (*model.User, error)
	GetAll() ([]*model.User, error)
	ClearCaches()
	InvalidateProfilesInChannelCacheByUser(userId string)
	InvalidateProfilesInChannelCache(channelId string)
	GetProfilesInChannel(options *model.UserGetOptions) ([]*model.User, error)
	GetProfilesInChannelByStatus(options *model.UserGetOptions) ([]*model.User, error)
	GetAllProfilesInChannel(channelId string, allowFromCache bool) (map[string]*model.User, error)
	GetProfilesNotInChannel(teamId string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetProfilesWithoutTeam(options *model.UserGetOptions) ([]*model.User, error)
	GetProfilesByUsernames(usernames []string, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetAllProfiles(options *model.UserGetOptions) ([]*model.User, error)
	GetProfiles(options *model.UserGetOptions) ([]*model.User, error)
	GetProfileByIds(userIds []string, options *UserGetByIdsOpts, allowFromCache bool) ([]*model.User, error)
	GetProfileByGroupChannelIdsForUser(userId string, channelIds []string) (map[string][]*model.User, error)
	InvalidateProfileCacheForUser(userId string)
	GetByEmail(email string) (*model.User, error)
	GetByAuth(authData *string, authService string) (*model.User, error)
	GetAllUsingAuthService(authService string) ([]*model.User, error)
	GetAllNotInAuthService(authServices []string) ([]*model.User, error)
	GetByUsername(username string) (*model.User, error)
	GetForLogin(loginId string, allowSignInWithUsername, allowSignInWithEmail bool) (*model.User, error)
	VerifyEmail(userId, email string) (string, error)
	GetEtagForAllProfiles() string
	GetEtagForProfiles(teamId string) string
	UpdateFailedPasswordAttempts(userId string, attempts int) error
	GetSystemAdminProfiles() (map[string]*model.User, error)
	PermanentDelete(userId string) error
	AnalyticsActiveCount(time int64, options model.UserCountOptions) (int64, error)
	AnalyticsActiveCountForPeriod(startTime int64, endTime int64, options model.UserCountOptions) (int64, error)
	GetUnreadCount(userId string) (int64, error)
	GetUnreadCountForChannel(userId string, channelId string) (int64, error)
	GetAnyUnreadPostCountForChannel(userId string, channelId string) (int64, error)
	GetRecentlyActiveUsersForTeam(teamId string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetNewUsersForTeam(teamId string, offset, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	Search(teamId string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchNotInTeam(notInTeamId string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchInChannel(channelId string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchNotInChannel(teamId string, channelId string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchWithoutTeam(term string, options *model.UserSearchOptions) ([]*model.User, error)
	SearchInGroup(groupID string, term string, options *model.UserSearchOptions) ([]*model.User, error)
	AnalyticsGetInactiveUsersCount() (int64, error)
	AnalyticsGetExternalUsers(hostDomain string) (bool, error)
	AnalyticsGetSystemAdminCount() (int64, error)
	AnalyticsGetGuestCount() (int64, error)
	GetProfilesNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, error)
	GetEtagForProfilesNotInTeam(teamId string) string
	ClearAllCustomRoleAssignments() error
	InferSystemInstallDate() (int64, error)
	GetAllAfter(limit int, afterId string) ([]*model.User, error)
	GetUsersBatchForIndexing(startTime, endTime int64, limit int) ([]*model.UserForIndexing, error)
	Count(options model.UserCountOptions) (int64, error)
	GetTeamGroupUsers(teamID string) ([]*model.User, error)
	GetChannelGroupUsers(channelID string) ([]*model.User, error)
	PromoteGuestToUser(userID string) error
	DemoteUserToGuest(userID string) error
	DeactivateGuests() ([]string, error)
	AutocompleteUsersInChannel(teamId, channelId, term string, options *model.UserSearchOptions) (*model.UserAutocompleteInChannel, error)
	GetKnownUsers(userID string) ([]string, error)
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
	Save(team *model.Team) (*model.Team, error)
	Update(team *model.Team) (*model.Team, error)
	Get(id string) (*model.Team, error)
	GetByName(name string) (*model.Team, error)
	GetByNames(name []string) ([]*model.Team, error)
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
	UserBelongsToTeams(userId string, teamIds []string) (bool, *model.AppError)
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
	CreateDirectChannel(userId *model.User, otherUserId *model.User) (*model.Channel, error)
	// SaveDirectChannel(channel *model.Channel, member1 *model.ChannelMember, member2 *model.ChannelMember) (*model.Channel, error)
	Update(channel *model.Channel) (*model.Channel, error)
	// UpdateSidebarChannelCategoryOnMove(channel *model.Channel, newTeamId string) error
	// ClearSidebarOnTeamLeave(userId, teamId string) error
	Get(id string, allowFromCache bool) (*model.Channel, error)
	InvalidateChannel(id string)
	InvalidateChannelByName(teamId, name string)
	// GetFromMaster(id string) (*model.Channel, error)
	// Delete(channelId string, time int64) error
	// Restore(channelId string, time int64) error
	// SetDeleteAt(channelId string, deleteAt int64, updateAt int64) error
	// PermanentDelete(channelId string) error
	// PermanentDeleteByTeam(teamId string) error
	GetByName(team_id string, name string, allowFromCache bool) (*model.Channel, error)
	GetByNames(team_id string, names []string, allowFromCache bool) ([]*model.Channel, error)
	GetByNameIncludeDeleted(team_id string, name string, allowFromCache bool) (*model.Channel, error)
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
	GetTeamChannels(teamId string) (*model.ChannelList, error)
	// GetAll(teamId string) ([]*model.Channel, error)
	// GetChannelsByIds(channelIds []string, includeDeleted bool) ([]*model.Channel, error)
	GetForPost(postId string) (*model.Channel, error)
	SaveMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, error)
	SaveMember(member *model.ChannelMember) (*model.ChannelMember, error)
	UpdateMember(member *model.ChannelMember) (*model.ChannelMember, error)
	UpdateMultipleMembers(members []*model.ChannelMember) ([]*model.ChannelMember, error)
	// GetMembers(channelId string, offset, limit int) (*model.ChannelMembers, *model.AppError)
	GetMember(channelId string, userId string) (*model.ChannelMember, error)
	// GetChannelMembersTimezones(channelId string) ([]model.StringMap, *model.AppError)
	GetAllChannelMembersForUser(userId string, allowFromCache bool, includeDeleted bool) (map[string]string, error)
	InvalidateAllChannelMembersForUser(userId string)
	// IsUserInChannelUseCache(userId string, channelId string) bool
	// GetAllChannelMembersNotifyPropsForChannel(channelId string, allowFromCache bool) (map[string]model.StringMap, *model.AppError)
	// InvalidateCacheForChannelMembersNotifyProps(channelId string)
	GetMemberForPost(postId string, userId string) (*model.ChannelMember, error)
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
	GetMembersForUser(teamId string, userId string) (*model.ChannelMembers, error)
	GetMembersForUserWithPagination(teamId, userId string, page, perPage int) (*model.ChannelMembers, error)
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
	UserBelongsToChannels(userId string, channelIds []string) (bool, *model.AppError)

	// // UpdateMembersRole sets all of the given team members to admins and all of the other members of the team to
	// // non-admin members.
	// UpdateMembersRole(channelID string, userIDs []string) *model.AppError

	// // GroupSyncedChannelCount returns the count of non-deleted group-constrained channels.
	// GroupSyncedChannelCount() (int64, *model.AppError)
}

type SchemeStore interface {
	Save(scheme *model.Scheme) (*model.Scheme, error)
	Get(schemeId string) (*model.Scheme, error)
	GetByName(schemeName string) (*model.Scheme, error)
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

type PreferenceStore interface {
	Save(preferences *model.Preferences) error
	// GetCategory(userId string, category string) (model.Preferences, error)
	Get(userId string, category string, name string) (*model.Preference, error)
	// GetAll(userId string) (model.Preferences, error)
	// Delete(userId, category, name string) error
	// DeleteCategory(userId string, category string) error
	// DeleteCategoryAndName(category string, name string) error
	// PermanentDeleteByUser(userId string) error
	// CleanupFlagsBatch(limit int64) (int64, error)
}

type UserGetByIdsOpts struct {
	// IsAdmin tracks whether or not the request is being made by an administrator. Does nothing when provided by a client.
	IsAdmin bool

	// Restrict to search in a list of teams and channels. Does nothing when provided by a client.
	ViewRestrictions *model.ViewUsersRestrictions

	// Since filters the users based on their UpdateAt timestamp.
	Since int64
}

type GroupStore interface {
	Create(group *model.Group) (*model.Group, *model.AppError)
	Get(groupID string) (*model.Group, *model.AppError)
	GetByName(name string, opts model.GroupSearchOpts) (*model.Group, *model.AppError)
	// GetByIDs(groupIDs []string) ([]*model.Group, *model.AppError)
	// GetByRemoteID(remoteID string, groupSource model.GroupSource) (*model.Group, *model.AppError)
	// GetAllBySource(groupSource model.GroupSource) ([]*model.Group, *model.AppError)
	// GetByUser(userId string) ([]*model.Group, *model.AppError)
	// Update(group *model.Group) (*model.Group, *model.AppError)
	// Delete(groupID string) (*model.Group, *model.AppError)

	// GetMemberUsers(groupID string) ([]*model.User, *model.AppError)
	// GetMemberUsersPage(groupID string, page int, perPage int) ([]*model.User, *model.AppError)
	// GetMemberCount(groupID string) (int64, *model.AppError)

	// GetMemberUsersInTeam(groupID string, teamID string) ([]*model.User, *model.AppError)
	// GetMemberUsersNotInChannel(groupID string, channelID string) ([]*model.User, *model.AppError)

	// UpsertMember(groupID string, userID string) (*model.GroupMember, *model.AppError)
	// DeleteMember(groupID string, userID string) (*model.GroupMember, *model.AppError)
	// PermanentDeleteMembersByUser(userId string) *model.AppError

	// CreateGroupSyncable(groupSyncable *model.GroupSyncable) (*model.GroupSyncable, *model.AppError)
	// GetGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, *model.AppError)
	// GetAllGroupSyncablesByGroupId(groupID string, syncableType model.GroupSyncableType) ([]*model.GroupSyncable, *model.AppError)
	// UpdateGroupSyncable(groupSyncable *model.GroupSyncable) (*model.GroupSyncable, *model.AppError)
	// DeleteGroupSyncable(groupID string, syncableID string, syncableType model.GroupSyncableType) (*model.GroupSyncable, *model.AppError)

	// // TeamMembersToAdd returns a slice of UserTeamIDPair that need newly created memberships
	// // based on the groups configurations. The returned list can be optionally scoped to a single given team.
	// //
	// // Typically since will be the last successful group sync time.
	// TeamMembersToAdd(since int64, teamID *string) ([]*model.UserTeamIDPair, *model.AppError)

	// // ChannelMembersToAdd returns a slice of UserChannelIDPair that need newly created memberships
	// // based on the groups configurations. The returned list can be optionally scoped to a single given channel.
	// //
	// // Typically since will be the last successful group sync time.
	// ChannelMembersToAdd(since int64, channelID *string) ([]*model.UserChannelIDPair, *model.AppError)

	// // TeamMembersToRemove returns all team members that should be removed based on group constraints.
	// TeamMembersToRemove(teamID *string) ([]*model.TeamMember, *model.AppError)

	// // ChannelMembersToRemove returns all channel members that should be removed based on group constraints.
	// ChannelMembersToRemove(channelID *string) ([]*model.ChannelMember, *model.AppError)

	// GetGroupsByChannel(channelId string, opts model.GroupSearchOpts) ([]*model.GroupWithSchemeAdmin, *model.AppError)
	// CountGroupsByChannel(channelId string, opts model.GroupSearchOpts) (int64, *model.AppError)

	// GetGroupsByTeam(teamId string, opts model.GroupSearchOpts) ([]*model.GroupWithSchemeAdmin, *model.AppError)
	// GetGroupsAssociatedToChannelsByTeam(teamId string, opts model.GroupSearchOpts) (map[string][]*model.GroupWithSchemeAdmin, *model.AppError)
	// CountGroupsByTeam(teamId string, opts model.GroupSearchOpts) (int64, *model.AppError)

	// GetGroups(page, perPage int, opts model.GroupSearchOpts) ([]*model.Group, *model.AppError)

	// TeamMembersMinusGroupMembers(teamID string, groupIDs []string, page, perPage int) ([]*model.UserWithGroups, *model.AppError)
	// CountTeamMembersMinusGroupMembers(teamID string, groupIDs []string) (int64, *model.AppError)
	// ChannelMembersMinusGroupMembers(channelID string, groupIDs []string, page, perPage int) ([]*model.UserWithGroups, *model.AppError)
	// CountChannelMembersMinusGroupMembers(channelID string, groupIDs []string) (int64, *model.AppError)

	// AdminRoleGroupsForSyncableMember returns the IDs of all of the groups that the user is a member of that are
	// configured as SchemeAdmin: true for the given syncable.
	AdminRoleGroupsForSyncableMember(userID, syncableID string, syncableType model.GroupSyncableType) ([]string, *model.AppError)

	// // PermittedSyncableAdmins returns the IDs of all of the user who are permitted by the group syncable to have
	// // the admin role for the given syncable.
	// PermittedSyncableAdmins(syncableID string, syncableType model.GroupSyncableType) ([]string, *model.AppError)

	// // GroupCount returns the total count of records in the UserGroups table.
	// GroupCount() (int64, *model.AppError)

	// // GroupTeamCount returns the total count of records in the GroupTeams table.
	// GroupTeamCount() (int64, *model.AppError)

	// // GroupChannelCount returns the total count of records in the GroupChannels table.
	// GroupChannelCount() (int64, *model.AppError)

	// // GroupMemberCount returns the total count of records in the GroupMembers table.
	// GroupMemberCount() (int64, *model.AppError)

	// // DistinctGroupMemberCount returns the count of records in the GroupMembers table with distinct UserId values.
	// DistinctGroupMemberCount() (int64, *model.AppError)

	// // GroupCountWithAllowReference returns the count of records in the Groups table with AllowReference set to true.
	// GroupCountWithAllowReference() (int64, *model.AppError)
}
