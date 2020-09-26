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

	//TODO: Open
	// if !user.IsGuest() {
	// 	userShouldBeAdmin, err := a.UserIsInAdminRoleGroup(user.Id, team.Id, model.GroupSyncableTypeTeam)
	// 	if err != nil {
	// 		return nil, false, err
	// 	}
	// 	tm.SchemeAdmin = userShouldBeAdmin
	// }

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
		return err
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
