package app

import (
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

// CreateDefaultChannels creates channels in the given team for each channel returned by (*App).DefaultChannelNames.
//
func (a *App) CreateDefaultChannels(teamID string) ([]*model.Channel, *model.AppError) {
	displayNames := map[string]string{
		"town-square": utils.T("api.channel.create_default_channels.town_square"),
		"off-topic":   utils.T("api.channel.create_default_channels.off_topic"),
	}
	channels := []*model.Channel{}
	defaultChannelNames := a.DefaultChannelNames()
	for _, name := range defaultChannelNames {
		displayName := utils.TDefault(displayNames[name], name)
		channel := &model.Channel{DisplayName: displayName, Name: name, Type: model.CHANNEL_OPEN, TeamId: teamID}
		if _, err := a.CreateChannel(channel, false); err != nil {
			return nil, err
		}
		channels = append(channels, channel)
	}
	return channels, nil
}

func (a *App) GetChannel(channelId string) (*model.Channel, *model.AppError) {
	channel, err := a.Srv().Store.Channel().Get(channelId, true)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannel", "app.channel.get.existing.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannel", "app.channel.get.find.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return channel, nil
}

func (a *App) addUserToChannel(user *model.User, channel *model.Channel, teamMember *model.TeamMember) (*model.ChannelMember, *model.AppError) {
	if channel.Type != model.CHANNEL_OPEN && channel.Type != model.CHANNEL_PRIVATE {
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user_to_channel.type.app_error", nil, "", http.StatusBadRequest)
	}

	channelMember, nErr := a.Srv().Store.Channel().GetMember(channel.Id, user.Id)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		if !errors.As(nErr, &nfErr) {
			return nil, model.NewAppError("AddUserToChannel", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	} else {
		return channelMember, nil
	}

	if channel.IsGroupConstrained() {
		nonMembers, err := a.FilterNonGroupChannelMembers([]string{user.Id}, channel)
		if err != nil {
			return nil, model.NewAppError("addUserToChannel", "api.channel.add_user_to_channel.type.app_error", nil, "", http.StatusInternalServerError)
		}
		if len(nonMembers) > 0 {
			return nil, model.NewAppError("addUserToChannel", "api.channel.add_members.user_denied", map[string]interface{}{"UserIDs": nonMembers}, "", http.StatusBadRequest)
		}
	}

	newMember := &model.ChannelMember{
		ChannelId:   channel.Id,
		UserId:      user.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
		SchemeGuest: user.IsGuest(),
		SchemeUser:  !user.IsGuest(),
	}

	if !user.IsGuest() {
		var userShouldBeAdmin bool
		userShouldBeAdmin, appErr := a.UserIsInAdminRoleGroup(user.Id, channel.Id, model.GroupSyncableTypeChannel)
		if appErr != nil {
			return nil, appErr
		}
		newMember.SchemeAdmin = userShouldBeAdmin
	}

	newMember, nErr = a.Srv().Store.Channel().SaveMember(newMember)
	if nErr != nil {
		mlog.Error("Failed to add member", mlog.String("user_id", user.Id), mlog.String("channel_id", channel.Id), mlog.Err(nErr))
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user.to.channel.failed.app_error", nil, "", http.StatusInternalServerError)
	}
	a.WaitForChannelMembership(channel.Id, user.Id)

	//TODO: Open
	// if nErr := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); nErr != nil {
	// 	mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(nErr))
	// 	return nil, model.NewAppError("AddUserToChannel", "app.channel_member_history.log_join_event.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
	// }

	//TODO: Open
	// a.InvalidateCacheForUser(user.Id)
	// a.invalidateCacheForChannelMembers(channel.Id)

	return newMember, nil
}

func (a *App) AddUserToChannel(user *model.User, channel *model.Channel) (*model.ChannelMember, *model.AppError) {
	teamMember, err := a.Srv().Store.Team().GetMember(channel.TeamId, user.Id)

	if err != nil {
		return nil, err
	}
	if teamMember.DeleteAt > 0 {
		return nil, model.NewAppError("AddUserToChannel", "api.channel.add_user.to.channel.failed.deleted.app_error", nil, "", http.StatusBadRequest)
	}

	newMember, err := a.addUserToChannel(user, channel, teamMember)
	if err != nil {
		return nil, err
	}

	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ADDED, "", channel.Id, "", nil)
	// message.Add("user_id", user.Id)
	// message.Add("team_id", channel.TeamId)
	// a.Publish(message)

	return newMember, nil
}

func (a *App) WaitForChannelMembership(channelId string, userId string) {
	if len(a.Config().SqlSettings.DataSourceReplicas) == 0 {
		return
	}

	now := model.GetMillis()

	for model.GetMillis()-now < 12000 {

		time.Sleep(100 * time.Millisecond)

		_, err := a.Srv().Store.Channel().GetMember(channelId, userId)

		// If the membership was found then return
		if err == nil {
			return
		}

		// If we received an error, but it wasn't a missing channel member then return
		var nfErr *store.ErrNotFound
		if !errors.As(err, &nfErr) {
			return
		}
	}

	mlog.Error("WaitForChannelMembership giving up", mlog.String("channel_id", channelId), mlog.String("user_id", userId))
}

func (a *App) JoinDefaultChannels(teamId string, user *model.User, shouldBeAdmin bool, userRequestorId string) *model.AppError {
  var requestor *model.User
  var nErr error
	if userRequestorId != "" {
    requestor, nErr = a.Srv().Store.User().Get(userRequestorId)
		if nErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(nErr, &nfErr):
				return model.NewAppError("JoinDefaultChannels", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
			default:
				return model.NewAppError("JoinDefaultChannels", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
	}

  var err *model.AppError
	for _, channelName := range a.DefaultChannelNames() {
		channel, channelErr := a.Srv().Store.Channel().GetByName(teamId, channelName, true)
		if channelErr != nil {
			var nfErr *store.ErrNotFound
			switch {
			case errors.As(err, &nfErr):
				err = model.NewAppError("JoinDefaultChannels", "app.channel.get_by_name.missing.app_error", nil, nfErr.Error(), http.StatusNotFound)
			default:
				err = model.NewAppError("JoinDefaultChannels", "app.channel.get_by_name.existing.app_error", nil, channelErr.Error(), http.StatusInternalServerError)
			}
			continue
		}

		if channel.Type != model.CHANNEL_OPEN {
			continue
		}

		cm := &model.ChannelMember{
			ChannelId:   channel.Id,
			UserId:      user.Id,
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
			SchemeAdmin: shouldBeAdmin,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}

		_, nErr = a.Srv().Store.Channel().SaveMember(cm)
		//TODO: Open
		// if histErr := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); histErr != nil {
		// 	mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(histErr))
		// 	return model.NewAppError("JoinDefaultChannels", "app.channel_member_history.log_join_event.internal_error", nil, histErr.Error(), http.StatusInternalServerError)
		// }

		//TODO: Open
		// if *a.Config().ServiceSettings.ExperimentalEnableDefaultChannelLeaveJoinMessages {
		// 	a.postJoinMessageForDefaultChannel(user, requestor, channel)
		// }

		//TODO: Open
		// a.invalidateCacheForChannelMembers(channel.Id)

		//TODO: Open
		// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ADDED, "", channel.Id, "", nil)
		// message.Add("user_id", user.Id)
		// message.Add("team_id", channel.TeamId)
		// a.Publish(message)

	}

	return err
}

// DefaultChannelNames returns the list of system-wide default channel names.
//
// By default the list will be (not necessarily in this order):
//	['town-square', 'off-topic']
// However, if TeamSettings.ExperimentalDefaultChannels contains a list of channels then that list will replace
// 'off-topic' and be included in the return results in addition to 'town-square'. For example:
//	['town-square', 'game-of-thrones', 'wow']
//
func (a *App) DefaultChannelNames() []string {
	names := []string{"town-square"}

	if len(a.Config().TeamSettings.ExperimentalDefaultChannels) == 0 {
		names = append(names, "off-topic")
	} else {
		seenChannels := map[string]bool{"town-square": true}
		for _, channelName := range a.Config().TeamSettings.ExperimentalDefaultChannels {
			if !seenChannels[channelName] {
				names = append(names, channelName)
				seenChannels[channelName] = true
			}
		}
	}

	return names
}

func (a *App) CreateChannel(channel *model.Channel, addMember bool) (*model.Channel, *model.AppError) {
	channel.DisplayName = strings.TrimSpace(channel.DisplayName)
	sc, nErr := a.Srv().Store.Channel().Save(channel, *a.Config().TeamSettings.MaxChannelsPerTeam)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.direct_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			return sc, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateChannel", "app.channel.create_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if addMember {
		user, err := a.Srv().Store.User().Get(channel.CreatorId)
		if err != nil {
			return nil, err
		}

		cm := &model.ChannelMember{
			ChannelId:   sc.Id,
			UserId:      user.Id,
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
			SchemeAdmin: true,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}

		if _, err := a.Srv().Store.Channel().SaveMember(cm); err != nil {
			return nil, err
		}
		//TODO: Open
		// if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(channel.CreatorId, sc.Id, model.GetMillis()); err != nil {
		// 	mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
		// 	return nil, model.NewAppError("CreateChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
		// }

		//TODO: Open
		// a.InvalidateCacheForUser(channel.CreatorId)
	}

	//TODO: Open
	// if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
	// 	a.Srv().Go(func() {
	// 		pluginContext := a.PluginContext()
	// 		pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	// 			hooks.ChannelHasBeenCreated(pluginContext, sc)
	// 			return true
	// 		}, plugin.ChannelHasBeenCreatedId)
	// 	})
	// }

	return sc, nil
}

func (a *App) CreateChannelWithUser(channel *model.Channel, userId string) (*model.Channel, *model.AppError) {
	if channel.IsGroupOrDirect() {
		return nil, model.NewAppError("CreateChannelWithUser", "api.channel.create_channel.direct_channel.app_error", nil, "", http.StatusBadRequest)
	}

	if len(channel.TeamId) == 0 {
		return nil, model.NewAppError("CreateChannelWithUser", "app.channel.create_channel.no_team_id.app_error", nil, "", http.StatusBadRequest)
	}

	// Get total number of channels on current team
	count, err := a.GetNumberOfChannelsOnTeam(channel.TeamId)
	if err != nil {
		return nil, err
	}

	if int64(count+1) > *a.Config().TeamSettings.MaxChannelsPerTeam {
		return nil, model.NewAppError("CreateChannelWithUser", "api.channel.create_channel.max_channel_limit.app_error", map[string]interface{}{"MaxChannelsPerTeam": *a.Config().TeamSettings.MaxChannelsPerTeam}, "", http.StatusBadRequest)
	}

	channel.CreatorId = userId

	rchannel, err := a.CreateChannel(channel, true)
	if err != nil {
		return nil, err
	}

	var user *model.User
	if user, err = a.GetUser(userId); err != nil {
		return nil, err
	}

	a.postJoinChannelMessage(user, channel)

	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_CREATED, "", "", userId, nil)
	// message.Add("channel_id", channel.Id)
	// message.Add("team_id", channel.TeamId)
	// a.Publish(message)

	return rchannel, nil
}

func (a *App) postJoinChannelMessage(user *model.User, channel *model.Channel) *model.AppError {
	//TODO: Open
	// message := fmt.Sprintf(utils.T("api.channel.join_channel.post_and_forget"), user.Username)
	// postType := model.POST_JOIN_CHANNEL

	// if user.IsGuest() {
	// 	message = fmt.Sprintf(utils.T("api.channel.guest_join_channel.post_and_forget"), user.Username)
	// 	postType = model.POST_GUEST_JOIN_CHANNEL
	// }

	// post := &model.Post{
	// 	ChannelId: channel.Id,
	// 	Message:   message,
	// 	Type:      postType,
	// 	UserId:    user.Id,
	// 	Props: model.StringInterface{
	// 		"username": user.Username,
	// 	},
	// }

	// if _, err := a.CreatePost(post, channel, false, true); err != nil {
	// 	return model.NewAppError("postJoinChannelMessage", "api.channel.post_user_add_remove_message_and_forget.error", nil, err.Error(), http.StatusInternalServerError)
	// }

	return nil
}

func (a *App) GetNumberOfChannelsOnTeam(teamId string) (int, *model.AppError) {
	// Get total number of channels on current team
	list, err := a.Srv().Store.Channel().GetTeamChannels(teamId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return 0, model.NewAppError("GetNumberOfChannelsOnTeam", "app.channel.get_channels.not_found.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return 0, model.NewAppError("GetNumberOfChannelsOnTeam", "app.channel.get_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return len(*list), nil
}

// UpdateChannel updates a given channel by its Id. It also publishes the CHANNEL_UPDATED event.
func (a *App) UpdateChannel(channel *model.Channel) (*model.Channel, *model.AppError) {
	_, err := a.Srv().Store.Channel().Update(channel)
	if err != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("UpdateChannel", "app.channel.update.bad_id", nil, invErr.Error(), http.StatusBadRequest)
		case errors.As(err, &appErr):
			return nil, appErr
		default:
			return nil, model.NewAppError("UpdateChannel", "app.channel.update_channel.internal_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	a.invalidateCacheForChannel(channel)

	//TODO: Open
	// messageWs := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_UPDATED, "", channel.Id, "", nil)
	// messageWs.Add("channel", channel.ToJson())
	// a.Publish(messageWs)

	return channel, nil
}

func (a *App) UpdateChannelMemberRoles(channelId string, userId string, newRoles string) (*model.ChannelMember, *model.AppError) {
	var member *model.ChannelMember
	var err *model.AppError
	if member, err = a.GetChannelMember(channelId, userId); err != nil {
		return nil, err
	}

	schemeGuestRole, schemeUserRole, schemeAdminRole, err := a.GetSchemeRolesForChannel(channelId)
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
				// If not part of the scheme for this channel, then it is not allowed to apply it as an explicit role.
				return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.scheme_role.app_error", nil, "role_name="+roleName, http.StatusBadRequest)
			}
		}
	}

	if member.SchemeUser && member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	if prevSchemeGuestValue != member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberRoles", "api.channel.update_channel_member_roles.changing_guest_role.app_error", nil, "", http.StatusBadRequest)
	}

	member.ExplicitRoles = strings.Join(newExplicitRoles, " ")

	member, nErr := a.Srv().Store.Channel().UpdateMember(member)
	if nErr != nil {
		var appErr *model.AppError
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("UpdateChannelMemberRoles", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("UpdateChannelMemberRoles", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	a.InvalidateCacheForUser(userId)
	return member, nil
}

func (a *App) UpdateChannelMemberSchemeRoles(channelId string, userId string, isSchemeGuest bool, isSchemeUser bool, isSchemeAdmin bool) (*model.ChannelMember, *model.AppError) {
	member, err := a.GetChannelMember(channelId, userId)
	if err != nil {
		return nil, err
	}

	member.SchemeAdmin = isSchemeAdmin
	member.SchemeUser = isSchemeUser
	member.SchemeGuest = isSchemeGuest

	if member.SchemeUser && member.SchemeGuest {
		return nil, model.NewAppError("UpdateChannelMemberSchemeRoles", "api.channel.update_channel_member_roles.guest_and_user.app_error", nil, "", http.StatusBadRequest)
	}

	// If the migration is not completed, we also need to check the default channel_admin/channel_user roles are not present in the roles field.
	if err = a.IsPhase2MigrationCompleted(); err != nil {
		member.ExplicitRoles = RemoveRoles([]string{model.CHANNEL_GUEST_ROLE_ID, model.CHANNEL_USER_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID}, member.ExplicitRoles)
	}

	member, nErr := a.Srv().Store.Channel().UpdateMember(member)
	if nErr != nil {
		var appErr *model.AppError
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("UpdateChannelMemberSchemeRoles", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("UpdateChannelMemberSchemeRoles", "app.channel.get_member.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	//TODO: Open
	// Notify the clients that the member notify props changed
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_CHANNEL_MEMBER_UPDATED, "", "", userId, nil)
	// message.Add("channelMember", member.ToJson())
	// a.Publish(message)

	a.InvalidateCacheForUser(userId)
	return member, nil
}

func (a *App) GetChannelMember(channelId string, userId string) (*model.ChannelMember, *model.AppError) {
	channelMember, err := a.Srv().Store.Channel().GetMember(channelId, userId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetChannelMember", MISSING_CHANNEL_MEMBER_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetChannelMember", "app.channel.get_member.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return channelMember, nil
}

// GetSchemeRolesForChannel Checks if a channel or its team has an override scheme for channel roles and returns the scheme roles or default channel roles.
func (a *App) GetSchemeRolesForChannel(channelId string) (guestRoleName, userRoleName, adminRoleName string, err *model.AppError) {
	channel, err := a.GetChannel(channelId)
	if err != nil {
		return
	}

	if channel.SchemeId != nil && len(*channel.SchemeId) != 0 {
		var scheme *model.Scheme
		scheme, err = a.GetScheme(*channel.SchemeId)
		if err != nil {
			return
		}

		guestRoleName = scheme.DefaultChannelGuestRole
		userRoleName = scheme.DefaultChannelUserRole
		adminRoleName = scheme.DefaultChannelAdminRole

		return
	}

	return a.GetTeamSchemeChannelRoles(channel.TeamId)
}

// GetTeamSchemeChannelRoles Checks if a team has an override scheme and returns the scheme channel role names or default channel role names.
func (a *App) GetTeamSchemeChannelRoles(teamId string) (guestRoleName, userRoleName, adminRoleName string, err *model.AppError) {
	team, err := a.GetTeam(teamId)
	if err != nil {
		return
	}

	if team.SchemeId != nil && len(*team.SchemeId) != 0 {
		var scheme *model.Scheme
		scheme, err = a.GetScheme(*team.SchemeId)
		if err != nil {
			return
		}

		guestRoleName = scheme.DefaultChannelGuestRole
		userRoleName = scheme.DefaultChannelUserRole
		adminRoleName = scheme.DefaultChannelAdminRole
	} else {
		guestRoleName = model.CHANNEL_GUEST_ROLE_ID
		userRoleName = model.CHANNEL_USER_ROLE_ID
		adminRoleName = model.CHANNEL_ADMIN_ROLE_ID
	}

	return
}

func (a *App) createDirectChannel(userId string, otherUserId string) (*model.Channel, *model.AppError) {
	uc1 := make(chan store.StoreResult, 1)
	uc2 := make(chan store.StoreResult, 1)
	go func() {
		user, err := a.Srv().Store.User().Get(userId)
		uc1 <- store.StoreResult{Data: user, Err: err}
		close(uc1)
	}()
	go func() {
		user, err := a.Srv().Store.User().Get(otherUserId)
		uc2 <- store.StoreResult{Data: user, Err: err}
		close(uc2)
	}()

	result := <-uc1
	if result.Err != nil {
		return nil, model.NewAppError("CreateDirectChannel", "api.channel.create_direct_channel.invalid_user.app_error", nil, userId, http.StatusBadRequest)
	}
	user := result.Data.(*model.User)

	result = <-uc2
	if result.Err != nil {
		return nil, model.NewAppError("CreateDirectChannel", "api.channel.create_direct_channel.invalid_user.app_error", nil, otherUserId, http.StatusBadRequest)
	}
	otherUser := result.Data.(*model.User)

	channel, nErr := a.Srv().Store.Channel().CreateDirectChannel(user, otherUser)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_direct_channel.not_direct.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("SqlChannelStore.Save", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			switch cErr.Resource {
			case "Channel":
				return channel, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
			case "ChannelMembers":
				return nil, model.NewAppError("CreateChannel", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
			}
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateDirectChannel", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	//TODO: Open
	// if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(userId, channel.Id, model.GetMillis()); err != nil {
	// 	mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
	// 	return nil, model.NewAppError("CreateDirectChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
	// }
	// if userId != otherUserId {
	// 	if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(otherUserId, channel.Id, model.GetMillis()); err != nil {
	// 		mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
	// 		return nil, model.NewAppError("CreateDirectChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
	// 	}
	// }

	return channel, nil
}

func (a *App) createGroupChannel(userIds []string, creatorId string) (*model.Channel, *model.AppError) {
	if len(userIds) > model.CHANNEL_GROUP_MAX_USERS || len(userIds) < model.CHANNEL_GROUP_MIN_USERS {
		return nil, model.NewAppError("CreateGroupChannel", "api.channel.create_group.bad_size.app_error", nil, "", http.StatusBadRequest)
	}

	users, err := a.Srv().Store.User().GetProfileByIds(userIds, nil, true)
	if err != nil {
		return nil, err
	}

	if len(users) != len(userIds) {
		return nil, model.NewAppError("CreateGroupChannel", "api.channel.create_group.bad_user.app_error", nil, "user_ids="+model.ArrayToJson(userIds), http.StatusBadRequest)
	}

	group := &model.Channel{
		Name:        model.GetGroupNameFromUserIds(userIds),
		DisplayName: model.GetGroupDisplayNameFromUsers(users, true),
		Type:        model.CHANNEL_GROUP,
	}

	channel, nErr := a.Srv().Store.Channel().Save(group, *a.Config().TeamSettings.MaxChannelsPerTeam)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		var cErr *store.ErrConflict
		var ltErr *store.ErrLimitExceeded
		var appErr *model.AppError
		switch {
		case errors.As(nErr, &invErr):
			switch {
			case invErr.Entity == "Channel" && invErr.Field == "DeleteAt":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.archived_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Type":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save.direct_channel.app_error", nil, "", http.StatusBadRequest)
			case invErr.Entity == "Channel" && invErr.Field == "Id":
				return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.existing.app_error", nil, "id="+invErr.Value.(string), http.StatusBadRequest)
			}
		case errors.As(nErr, &cErr):
			return channel, model.NewAppError("CreateChannel", store.CHANNEL_EXISTS_ERROR, nil, cErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &ltErr):
			return nil, model.NewAppError("CreateChannel", "store.sql_channel.save_channel.limit.app_error", nil, ltErr.Error(), http.StatusBadRequest)
		case errors.As(nErr, &appErr): // in case we haven't converted to plain error.
			return nil, appErr
		default: // last fallback in case it doesn't map to an existing app error.
			return nil, model.NewAppError("CreateChannel", "app.channel.create_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	for _, user := range users {
		cm := &model.ChannelMember{
			UserId:      user.Id,
			ChannelId:   group.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: user.IsGuest(),
			SchemeUser:  !user.IsGuest(),
		}

		if _, nErr = a.Srv().Store.Channel().SaveMember(cm); nErr != nil {
			var appErr *model.AppError
			var cErr *store.ErrConflict
			switch {
			case errors.As(nErr, &cErr):
				switch cErr.Resource {
				case "ChannelMembers":
					return nil, model.NewAppError("createGroupChannel", "app.channel.save_member.exists.app_error", nil, cErr.Error(), http.StatusBadRequest)
				}
			case errors.As(nErr, &appErr):
				return nil, appErr
			default:
				return nil, model.NewAppError("createGroupChannel", "app.channel.create_direct_channel.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
			}
		}
		//TODO: Open
		// if err := a.Srv().Store.ChannelMemberHistory().LogJoinEvent(user.Id, channel.Id, model.GetMillis()); err != nil {
		// 	mlog.Error("Failed to update ChannelMemberHistory table", mlog.Err(err))
		// 	return nil, model.NewAppError("createGroupChannel", "app.channel_member_history.log_join_event.internal_error", nil, err.Error(), http.StatusInternalServerError)
		// }
	}

	return channel, nil
}
