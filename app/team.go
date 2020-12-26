package app

import (
	"errors"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"net/http"
	"strings"
)

func (a *App) GetTeamMember(teamId, userId string) (*model.TeamMember, *model.AppError) {
	return a.Srv().Store.Team().GetMember(teamId, userId)
}

// Returns three values:
// 1. a pointer to the team member, if successful
// 2. a boolean: true if the user has a non-deleted team member for that team already, otherwise false.
// 3. a pointer to an AppError if something went wrong.
func (a *App) joinUserToTeam(team *model.Team, user *model.User) (*model.TeamMember, bool, *model.AppError) {
	tm := &model.TeamMember{
		TeamId:      team.Id,
		UserId:      user.Id,
		SchemeGuest: user.IsGuest(),
		SchemeUser:  !user.IsGuest(),
	}

	if !user.IsGuest() {
		userShouldBeAdmin, err := a.UserIsInAdminRoleGroup(user.Id, team.Id, model.GroupSyncableTypeTeam)
		if err != nil {
			return nil, false, err
		}
		tm.SchemeAdmin = userShouldBeAdmin
	}

	if team.Email == user.Email {
		tm.SchemeAdmin = true
	}

	rtm, err := a.Srv().Store.Team().GetMember(team.Id, user.Id)
	if err != nil {
		// Membership appears to be missing. Lets try to add.
		tmr, nErr := a.Srv().Store.Team().SaveMember(tm, *a.Config().TeamSettings.MaxUsersPerTeam)
		if nErr != nil {
			var appErr *model.AppError
			var conflictErr *store.ErrConflict
			var limitExeededErr *store.ErrLimitExceeded
			switch {
			case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
				return nil, false, appErr
			case errors.As(nErr, &conflictErr):
				return nil, false, model.NewAppError("joinUserToTeam", "app.team.join_user_to_team.save_member.conflict.app_error", nil, nErr.Error(), http.StatusBadRequest)
			case errors.As(nErr, &limitExeededErr):
				return nil, false, model.NewAppError("joinUserToTeam", "app.team.join_user_to_team.save_member.max_accounts.app_error", nil, nErr.Error(), http.StatusBadRequest)
			default: // last fallback in case it doesn't map to an existing app error.
				return nil, false, model.NewAppError("joinUserToTeam", "app.team.join_user_to_team.save_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
		return tmr, false, nil
	}

	// Membership already exists.  Check if deleted and update, otherwise do nothing
	// Do nothing if already added
	if rtm.DeleteAt == 0 {
		return rtm, true, nil
	}

	membersCount, err := a.Srv().Store.Team().GetActiveMemberCount(tm.TeamId, nil)
	if err != nil {
		return nil, false, err
	}

	if membersCount >= int64(*a.Config().TeamSettings.MaxUsersPerTeam) {
		return nil, false, model.NewAppError("joinUserToTeam", "app.team.join_user_to_team.max_accounts.app_error", nil, "teamId="+tm.TeamId, http.StatusBadRequest)
	}

	member, err := a.Srv().Store.Team().UpdateMember(tm)
	if err != nil {
		return nil, false, err
	}

	return member, false, nil
}

func (a *App) JoinUserToTeam(team *model.Team, user *model.User, userRequestorId string) *model.AppError {
	if !a.isTeamEmailAllowed(user, team) {
		return model.NewAppError("JoinUserToTeam", "api.team.join_user_to_team.allowed_domains.app_error", nil, "", http.StatusBadRequest)
	}
	//tm
	_, alreadyAdded, err := a.joinUserToTeam(team, user)
	if err != nil {
		return err
	}
	if alreadyAdded {
		return nil
	}

	//TODO: Open
	// if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
	// 	var actor *model.User
	// 	if userRequestorId != "" {
	// 		actor, _ = a.GetUser(userRequestorId)
	// 	}

	// 	a.Srv().Go(func() {
	// 		pluginContext := a.PluginContext()
	// 		pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	// 			hooks.UserHasJoinedTeam(pluginContext, tm, actor)
	// 			return true
	// 		}, plugin.UserHasJoinedTeamId)
	// 	})
	// }

	if _, err := a.Srv().Store.User().UpdateUpdateAt(user.Id); err != nil {
		return model.NewAppError("JoinUserToTeam", "app.user.update_update.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := a.createInitialSidebarCategories(user.Id, team.Id); err != nil {
		mlog.Error(
			"Encountered an issue creating default sidebar categories.",
			mlog.String("user_id", user.Id),
			mlog.String("team_id", team.Id),
			mlog.Err(err),
		)
	}

	shouldBeAdmin := team.Email == user.Email

	if !user.IsGuest() {
		// Soft error if there is an issue joining the default channels
		if err := a.JoinDefaultChannels(team.Id, user, shouldBeAdmin, userRequestorId); err != nil {
			mlog.Error(
				"Encountered an issue joining default channels.",
				mlog.String("user_id", user.Id),
				mlog.String("team_id", team.Id),
				mlog.Err(err),
			)
		}
	}

	//TODO: Open
	// a.ClearSessionCacheForUser(user.Id)
	// a.InvalidateCacheForUser(user.Id)
	// a.invalidateCacheForUserTeams(user.Id)

	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_ADDED_TO_TEAM, "", "", user.Id, nil)
	// message.Add("team_id", team.Id)
	// message.Add("user_id", user.Id)
	// a.Publish(message)

	return nil
}

func (a *App) isTeamEmailAllowed(user *model.User, team *model.Team) bool {
	if user.IsBot {
		return true
	}
	email := strings.ToLower(user.Email)
	allowedDomains := a.getAllowedDomains(user, team)
	return a.isEmailAddressAllowed(email, allowedDomains)
}

func (a *App) getAllowedDomains(user *model.User, team *model.Team) []string {
	if user.IsGuest() {
		return []string{*a.Config().GuestAccountsSettings.RestrictCreationToDomains}
	}
	// First check per team allowedDomains, then app wide restrictions
	return []string{team.AllowedDomains, *a.Config().TeamSettings.RestrictCreationToDomains}
}

func (a *App) isEmailAddressAllowed(email string, allowedDomains []string) bool {
	for _, restriction := range allowedDomains {
		domains := a.normalizeDomains(restriction)
		if len(domains) <= 0 {
			continue
		}
		matched := false
		for _, d := range domains {
			if strings.HasSuffix(email, "@"+d) {
				matched = true
				break
			}
		}
		if !matched {
			return false
		}
	}

	return true
}

func (a *App) normalizeDomains(domains string) []string {
	// commas and @ signs are optional
	// can be in the form of "@corp.mattermost.com, mattermost.com mattermost.org" -> corp.mattermost.com mattermost.com mattermost.org
	return strings.Fields(strings.TrimSpace(strings.ToLower(strings.Replace(strings.Replace(domains, "@", " ", -1), ",", " ", -1))))
}

func (a *App) CreateTeam(team *model.Team) (*model.Team, *model.AppError) {
	team.InviteId = ""
	rteam, err := a.Srv().Store.Team().Save(team)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("CreateTeam", "app.team.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("CreateTeam", "app.team.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if _, err := a.CreateDefaultChannels(rteam.Id); err != nil {
		return nil, err
	}

	return rteam, nil
}

func (a *App) CreateTeamWithUser(team *model.Team, userId string) (*model.Team, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}
	team.Email = user.Email

	if !a.isTeamEmailAllowed(user, team) {
		return nil, model.NewAppError("isTeamEmailAllowed", "api.team.is_team_creation_allowed.domain.app_error", nil, "", http.StatusBadRequest)
	}

	rteam, err := a.CreateTeam(team)
	if err != nil {
		return nil, err
	}

	if err = a.JoinUserToTeam(rteam, user, ""); err != nil {
		return nil, err
	}

	return rteam, nil
}

func (a *App) updateTeamUnsanitized(team *model.Team) (*model.Team, *model.AppError) {
	team, err := a.Srv().Store.Team().Update(team)
	if err != nil {
		var invErr *store.ErrInvalidInput
		var appErr *model.AppError
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("updateTeamUnsanitized", "app.team.update.find.app_error", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("updateTeamUnsanitized", "app.team.update.updating.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return team, nil
}

func (a *App) UpdateTeamMemberRoles(teamId string, userId string, newRoles string) (*model.TeamMember, *model.AppError) {
	member, nErr := a.Srv().Store.Team().GetMember(teamId, userId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("UpdateTeamMemberRoles", "app.team.get_member.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("UpdateTeamMemberRoles", "app.team.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if member == nil {
		return nil, model.NewAppError("UpdateTeamMemberRoles", "api.team.update_member_roles.not_a_member", nil, "userId="+userId+" teamId="+teamId, http.StatusBadRequest)
	}

	schemeGuestRole, schemeUserRole, schemeAdminRole, err := a.GetSchemeRolesForTeam(teamId)
	if err != nil {
		return nil, err
	}

	prevSchemeGuestValue := member.SchemeGuest

	var newExplicitRoles []string
	member.SchemeGuest = false
	member.SchemeUser = false
	member.SchemeAdmin = false

	for _, roleName := range strings.Fields(newRoles) {
		var role *model.Role
		role, err = a.GetRoleByName(roleName)
		if err != nil {
			err.StatusCode = http.StatusBadRequest
			return nil, err
		}
		if !role.SchemeManaged {
			// The role is not scheme-managed, so it's OK to apply it to the explicit roles field.
			newExplicitRoles = append(newExplicitRoles, roleName)
		} else {
			// The role is scheme-managed, so need to check if it is part of the scheme for this channel or not.
			switch roleName {
			case schemeAdminRole:
				member.SchemeAdmin = true
			case schemeUserRole:
				member.SchemeUser = true
			case schemeGuestRole:
				member.SchemeGuest = true
			default:
				// If not part of the scheme for this team, then it is not allowed to apply it as an explicit role.
				return nil, model.NewAppError("UpdateTeamMemberRoles", "api.channel.update_team_member_roles.scheme_role.app_error", nil, "role_name="+roleName, http.StatusBadRequest)
			}
		}
	}

	if member.SchemeGuest && member.SchemeUser {
		return nil, model.NewAppError("UpdateTeamMemberRoles", "api.team.update_team_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	if prevSchemeGuestValue != member.SchemeGuest {
		return nil, model.NewAppError("UpdateTeamMemberRoles", "api.channel.update_team_member_roles.changing_guest_role.app_error", nil, "", http.StatusBadRequest)
	}

	member.ExplicitRoles = strings.Join(newExplicitRoles, " ")

	member, nErr = a.Srv().Store.Team().UpdateMember(member)
	if nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("UpdateTeamMemberRoles", "app.team.save_member.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	a.ClearSessionCacheForUser(userId)

	a.sendUpdatedMemberRoleEvent(userId, member)

	return member, nil
}

func (a *App) UpdateTeamMemberSchemeRoles(teamId string, userId string, isSchemeGuest bool, isSchemeUser bool, isSchemeAdmin bool) (*model.TeamMember, *model.AppError) {
	member, err := a.GetTeamMember(teamId, userId)
	if err != nil {
		return nil, err
	}

	member.SchemeAdmin = isSchemeAdmin
	member.SchemeUser = isSchemeUser
	member.SchemeGuest = isSchemeGuest

	if member.SchemeUser && member.SchemeGuest {
		return nil, model.NewAppError("UpdateTeamMemberSchemeRoles", "api.team.update_team_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	// If the migration is not completed, we also need to check the default team_admin/team_user roles are not present in the roles field.
	if err = a.IsPhase2MigrationCompleted(); err != nil {
		member.ExplicitRoles = RemoveRoles([]string{model.TEAM_GUEST_ROLE_ID, model.TEAM_USER_ROLE_ID, model.TEAM_ADMIN_ROLE_ID}, member.ExplicitRoles)
	}

	member, nErr := a.Srv().Store.Team().UpdateMember(member)
	if nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("UpdateTeamMemberSchemeRoles", "app.team.save_member.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	a.ClearSessionCacheForUser(userId)

	a.sendUpdatedMemberRoleEvent(userId, member)

	return member, nil
}

func (a *App) GetTeam(teamId string) (*model.Team, *model.AppError) {
	team, err := a.Srv().Store.Team().Get(teamId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetTeam", "app.team.get.find.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetTeam", "app.team.get.finding.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return team, nil
}

func (a *App) GetSchemeRolesForTeam(teamId string) (string, string, string, *model.AppError) {
	team, err := a.GetTeam(teamId)
	if err != nil {
		return "", "", "", err
	}

	if team.SchemeId != nil && len(*team.SchemeId) != 0 {
		scheme, err := a.GetScheme(*team.SchemeId)
		if err != nil {
			return "", "", "", err
		}
		return scheme.DefaultTeamGuestRole, scheme.DefaultTeamUserRole, scheme.DefaultTeamAdminRole, nil
	}

	return model.TEAM_GUEST_ROLE_ID, model.TEAM_USER_ROLE_ID, model.TEAM_ADMIN_ROLE_ID, nil
}

func (a *App) sendUpdatedMemberRoleEvent(userId string, member *model.TeamMember) {
	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_MEMBERROLE_UPDATED, "", "", userId, nil)
	// message.Add("member", member.ToJson())
	// a.Publish(message)
}

func (a *App) RemoveTeamMemberFromTeam(teamMember *model.TeamMember, requestorId string) *model.AppError {
	// Send the websocket message before we actually do the remove so the user being removed gets it.
	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_LEAVE_TEAM, teamMember.TeamId, "", "", nil)
	message.Add("user_id", teamMember.UserId)
	message.Add("team_id", teamMember.TeamId)
	a.Publish(message)

	user, nErr := a.Srv().Store.User().Get(teamMember.UserId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return model.NewAppError("RemoveTeamMemberFromTeam", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return model.NewAppError("RemoveTeamMemberFromTeam", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	teamMember.Roles = ""
	teamMember.DeleteAt = model.GetMillis()

	if _, nErr := a.Srv().Store.Team().UpdateMember(teamMember); nErr != nil {
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &appErr):
			return appErr
		default:
			return model.NewAppError("RemoveTeamMemberFromTeam", "app.team.save_member.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
		var actor *model.User
		if requestorId != "" {
			actor, _ = a.GetUser(requestorId)
		}

		a.Srv().Go(func() {
			pluginContext := a.PluginContext()
			pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
				hooks.UserHasLeftTeam(pluginContext, teamMember, actor)
				return true
			}, plugin.UserHasLeftTeamId)
		})
	}

	if _, err := a.Srv().Store.User().UpdateUpdateAt(user.Id); err != nil {
		return model.NewAppError("RemoveTeamMemberFromTeam", "app.user.update_update.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	if err := a.Srv().Store.Channel().ClearSidebarOnTeamLeave(user.Id, teamMember.TeamId); err != nil {
		return model.NewAppError("RemoveTeamMemberFromTeam", "app.channel.sidebar_categories.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	// delete the preferences that set the last channel used in the team and other team specific preferences
	if err := a.Srv().Store.Preference().DeleteCategory(user.Id, teamMember.TeamId); err != nil {
		return model.NewAppError("RemoveTeamMemberFromTeam", "app.preference.delete.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	a.ClearSessionCacheForUser(user.Id)
	a.InvalidateCacheForUser(user.Id)
	a.invalidateCacheForUserTeams(user.Id)

	return nil
}

func (a *App) GetTeamMembersByIds(teamId string, userIds []string, restrictions *model.ViewUsersRestrictions) ([]*model.TeamMember, *model.AppError) {
	teamMembers, err := a.Srv().Store.Team().GetMembersByIds(teamId, userIds, restrictions)
	if err != nil {
		return nil, model.NewAppError("GetTeamMembersByIds", "app.team.get_members_by_ids.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return teamMembers, nil
}
