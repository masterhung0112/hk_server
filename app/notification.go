package app

import (
	"github.com/masterhung0112/hk_server/model"
)

func (a *App) SendNotifications(post *model.Post, team *model.Team, channel *model.Channel, sender *model.User, parentPostList *model.PostList, setOnline bool) ([]string, error) {
	return []string{}, nil
	// // Do not send notifications in archived channels
	// if channel.DeleteAt > 0 {
	// 	return []string{}, nil
	// }

	// pchan := make(chan store.StoreResult, 1)
	// go func() {
	// 	props, err := a.Srv().Store.User().GetAllProfilesInChannel(channel.Id, true)
	// 	pchan <- store.StoreResult{Data: props, NErr: err}
	// 	close(pchan)
	// }()

	// cmnchan := make(chan store.StoreResult, 1)
	// go func() {
	// 	props, err := a.Srv().Store.Channel().GetAllChannelMembersNotifyPropsForChannel(channel.Id, true)
	// 	cmnchan <- store.StoreResult{Data: props, NErr: err}
	// 	close(cmnchan)
	// }()

	// var gchan chan store.StoreResult
	// if a.allowGroupMentions(post) {
	// 	gchan = make(chan store.StoreResult, 1)
	// 	go func() {
	// 		groupsMap, err := a.getGroupsAllowedForReferenceInChannel(channel, team)
	// 		gchan <- store.StoreResult{Data: groupsMap, Err: err}
	// 		close(gchan)
	// 	}()
	// }

	// var fchan chan store.StoreResult
	// if len(post.FileIds) != 0 {
	// 	fchan = make(chan store.StoreResult, 1)
	// 	go func() {
	// 		fileInfos, err := a.Srv().Store.FileInfo().GetForPost(post.Id, true, false, true)
	// 		fchan <- store.StoreResult{Data: fileInfos, NErr: err}
	// 		close(fchan)
	// 	}()
	// }

	// result := <-pchan
	// if result.NErr != nil {
	// 	return nil, result.NErr
	// }
	// profileMap := result.Data.(map[string]*model.User)

	// result = <-cmnchan
	// if result.NErr != nil {
	// 	return nil, result.NErr
	// }
	// channelMemberNotifyPropsMap := result.Data.(map[string]model.StringMap)

	// groups := make(map[string]*model.Group)
	// if gchan != nil {
	// 	result = <-gchan
	// 	if result.Err != nil {
	// 		return nil, result.Err
	// 	}
	// 	groups = result.Data.(map[string]*model.Group)
	// }

	// mentions := &ExplicitMentions{}
	// allActivityPushUserIds := []string{}

	// if channel.Type == model.CHANNEL_DIRECT {
	// 	otherUserId := channel.GetOtherUserIdForDM(post.UserId)

	// 	_, ok := profileMap[otherUserId]
	// 	if ok {
	// 		mentions.addMention(otherUserId, DMMention)
	// 	}

	// 	if post.GetProp("from_webhook") == "true" {
	// 		mentions.addMention(post.UserId, DMMention)
	// 	}
	// } else {
	// 	allowChannelMentions := a.allowChannelMentions(post, len(profileMap))
	// 	keywords := a.getMentionKeywordsInChannel(profileMap, allowChannelMentions, channelMemberNotifyPropsMap)

	// 	mentions = getExplicitMentions(post, keywords, groups)

	// 	// Add an implicit mention when a user is added to a channel
	// 	// even if the user has set 'username mentions' to false in account settings.
	// 	if post.Type == model.POST_ADD_TO_CHANNEL {
	// 		addedUserId, ok := post.GetProp(model.POST_PROPS_ADDED_USER_ID).(string)
	// 		if ok {
	// 			mentions.addMention(addedUserId, KeywordMention)
	// 		}
	// 	}

	// 	// Iterate through all groups that were mentioned and insert group members into the list of mentions or potential mentions
	// 	for _, group := range mentions.GroupMentions {
	// 		anyUsersMentionedByGroup, err := a.insertGroupMentions(group, channel, profileMap, mentions)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		if !anyUsersMentionedByGroup {
	// 			a.sendNoUsersNotifiedByGroupInChannel(sender, post, channel, group)
	// 		}
	// 	}

	// 	// get users that have comment thread mentions enabled
	// 	if len(post.RootId) > 0 && parentPostList != nil {
	// 		for _, threadPost := range parentPostList.Posts {
	// 			profile := profileMap[threadPost.UserId]
	// 			if profile != nil && (profile.NotifyProps[model.COMMENTS_NOTIFY_PROP] == model.COMMENTS_NOTIFY_ANY || (profile.NotifyProps[model.COMMENTS_NOTIFY_PROP] == model.COMMENTS_NOTIFY_ROOT && threadPost.Id == parentPostList.Order[0])) {
	// 				mentionType := ThreadMention
	// 				if threadPost.Id == parentPostList.Order[0] {
	// 					mentionType = CommentMention
	// 				}

	// 				mentions.addMention(threadPost.UserId, mentionType)
	// 			}
	// 		}
	// 	}

	// 	// prevent the user from mentioning themselves
	// 	if post.GetProp("from_webhook") != "true" {
	// 		mentions.removeMention(post.UserId)
	// 	}

	// 	go func() {
	// 		_, err := a.sendOutOfChannelMentions(sender, post, channel, mentions.OtherPotentialMentions)
	// 		if err != nil {
	// 			mlog.Error("Failed to send warning for out of channel mentions", mlog.String("user_id", sender.Id), mlog.String("post_id", post.Id), mlog.Err(err))
	// 		}
	// 	}()

	// 	// find which users in the channel are set up to always receive mobile notifications
	// 	for _, profile := range profileMap {
	// 		if (profile.NotifyProps[model.PUSH_NOTIFY_PROP] == model.USER_NOTIFY_ALL ||
	// 			channelMemberNotifyPropsMap[profile.Id][model.PUSH_NOTIFY_PROP] == model.CHANNEL_NOTIFY_ALL) &&
	// 			(post.UserId != profile.Id || post.GetProp("from_webhook") == "true") &&
	// 			!post.IsSystemMessage() {
	// 			allActivityPushUserIds = append(allActivityPushUserIds, profile.Id)
	// 		}
	// 	}
	// }

	// mentionedUsersList := make([]string, 0, len(mentions.Mentions))
	// updateMentionChans := []chan *model.AppError{}
	// mentionAutofollowChans := []chan *model.AppError{}
	// threadParticipants := []string{post.UserId}
	// if *a.Config().ServiceSettings.ThreadAutoFollow && post.RootId != "" {
	// 	if parentPostList != nil {
	// 		threadParticipants = append(threadParticipants, parentPostList.Posts[parentPostList.Order[0]].UserId)
	// 	}
	// 	for id := range mentions.Mentions {
	// 		threadParticipants = append(threadParticipants, id)
	// 	}
	// 	// for each mention, make sure to update thread autofollow
	// 	for _, id := range threadParticipants {
	// 		mac := make(chan *model.AppError, 1)
	// 		go func(userId string) {
	// 			defer close(mac)

	// 			nErr := a.Srv().Store.Thread().CreateMembershipIfNeeded(userId, post.RootId, true)
	// 			if nErr != nil {
	// 				mac <- model.NewAppError("SendNotifications", "app.channel.autofollow.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	// 				return
	// 			}

	// 			mac <- nil
	// 		}(id)
	// 		mentionAutofollowChans = append(mentionAutofollowChans, mac)
	// 	}
	// }
	// for id := range mentions.Mentions {
	// 	mentionedUsersList = append(mentionedUsersList, id)

	// 	umc := make(chan *model.AppError, 1)
	// 	go func(userId string) {
	// 		defer close(umc)
	// 		nErr := a.Srv().Store.Channel().IncrementMentionCount(post.ChannelId, userId, *a.Config().ServiceSettings.ThreadAutoFollow)
	// 		if nErr != nil {
	// 			umc <- model.NewAppError("SendNotifications", "app.channel.increment_mention_count.app_error", nil, nErr.Error(), http.StatusInternalServerError)
	// 			return
	// 		}
	// 		umc <- nil
	// 	}(id)
	// 	updateMentionChans = append(updateMentionChans, umc)
	// }

	// notification := &PostNotification{
	// 	Post:       post,
	// 	Channel:    channel,
	// 	ProfileMap: profileMap,
	// 	Sender:     sender,
	// }

	// if *a.Config().EmailSettings.SendEmailNotifications {
	// 	for _, id := range mentionedUsersList {
	// 		if profileMap[id] == nil {
	// 			continue
	// 		}

	// 		//If email verification is required and user email is not verified don't send email.
	// 		if *a.Config().EmailSettings.RequireEmailVerification && !profileMap[id].EmailVerified {
	// 			mlog.Error("Skipped sending notification email, address not verified.", mlog.String("user_email", profileMap[id].Email), mlog.String("user_id", id))
	// 			continue
	// 		}

	// 		if a.userAllowsEmail(profileMap[id], channelMemberNotifyPropsMap[id], post) {
	// 			a.sendNotificationEmail(notification, profileMap[id], team)
	// 		}
	// 	}
	// }

	// // Check for channel-wide mentions in channels that have too many members for those to work
	// if int64(len(profileMap)) > *a.Config().TeamSettings.MaxNotificationsPerChannel {
	// 	T := utils.GetUserTranslations(sender.Locale)

	// 	if mentions.HereMentioned {
	// 		a.SendEphemeralPost(
	// 			post.UserId,
	// 			&model.Post{
	// 				ChannelId: post.ChannelId,
	// 				Message:   T("api.post.disabled_here", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
	// 				CreateAt:  post.CreateAt + 1,
	// 			},
	// 		)
	// 	}

	// 	if mentions.ChannelMentioned {
	// 		a.SendEphemeralPost(
	// 			post.UserId,
	// 			&model.Post{
	// 				ChannelId: post.ChannelId,
	// 				Message:   T("api.post.disabled_channel", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
	// 				CreateAt:  post.CreateAt + 1,
	// 			},
	// 		)
	// 	}

	// 	if mentions.AllMentioned {
	// 		a.SendEphemeralPost(
	// 			post.UserId,
	// 			&model.Post{
	// 				ChannelId: post.ChannelId,
	// 				Message:   T("api.post.disabled_all", map[string]interface{}{"Users": *a.Config().TeamSettings.MaxNotificationsPerChannel}),
	// 				CreateAt:  post.CreateAt + 1,
	// 			},
	// 		)
	// 	}
	// }

	// // Make sure all mention updates are complete to prevent race
	// // Probably better to batch these DB updates in the future
	// // MUST be completed before push notifications send
	// for _, umc := range updateMentionChans {
	// 	if err := <-umc; err != nil {
	// 		mlog.Warn(
	// 			"Failed to update mention count",
	// 			mlog.String("post_id", post.Id),
	// 			mlog.String("channel_id", post.ChannelId),
	// 			mlog.Err(err),
	// 		)
	// 	}
	// }

	// // Log the problems that might have occurred while auto following the thread
	// for _, mac := range mentionAutofollowChans {
	// 	if err := <-mac; err != nil {
	// 		mlog.Warn(
	// 			"Failed to update thread autofollow from mention",
	// 			mlog.String("post_id", post.Id),
	// 			mlog.String("channel_id", post.ChannelId),
	// 			mlog.Err(err),
	// 		)
	// 	}
	// }
	// sendPushNotifications := false
	// if *a.Config().EmailSettings.SendPushNotifications {
	// 	pushServer := *a.Config().EmailSettings.PushNotificationServer
	// 	if license := a.Srv().License(); pushServer == model.MHPNS && (license == nil || !*license.Features.MHPNS) {
	// 		mlog.Warn("Push notifications are disabled. Go to System Console > Notifications > Mobile Push to enable them.")
	// 		sendPushNotifications = false
	// 	} else {
	// 		sendPushNotifications = true
	// 	}
	// }

	// if sendPushNotifications {
	// 	for _, id := range mentionedUsersList {
	// 		if profileMap[id] == nil {
	// 			continue
	// 		}

	// 		var status *model.Status
	// 		var err *model.AppError
	// 		if status, err = a.GetStatus(id); err != nil {
	// 			status = &model.Status{UserId: id, Status: model.STATUS_OFFLINE, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
	// 		}

	// 		if ShouldSendPushNotification(profileMap[id], channelMemberNotifyPropsMap[id], true, status, post) {
	// 			mentionType := mentions.Mentions[id]

	// 			replyToThreadType := ""
	// 			if mentionType == ThreadMention {
	// 				replyToThreadType = model.COMMENTS_NOTIFY_ANY
	// 			} else if mentionType == CommentMention {
	// 				replyToThreadType = model.COMMENTS_NOTIFY_ROOT
	// 			}

	// 			a.sendPushNotification(
	// 				notification,
	// 				profileMap[id],
	// 				mentionType == KeywordMention || mentionType == ChannelMention || mentionType == DMMention,
	// 				mentionType == ChannelMention,
	// 				replyToThreadType,
	// 			)
	// 		} else {
	// 			// register that a notification was not sent
	// 			a.NotificationsLog().Warn("Notification not sent",
	// 				mlog.String("ackId", ""),
	// 				mlog.String("type", model.PUSH_TYPE_MESSAGE),
	// 				mlog.String("userId", id),
	// 				mlog.String("postId", post.Id),
	// 				mlog.String("status", model.PUSH_NOT_SENT),
	// 			)
	// 		}
	// 	}

	// 	for _, id := range allActivityPushUserIds {
	// 		if profileMap[id] == nil {
	// 			continue
	// 		}

	// 		if _, ok := mentions.Mentions[id]; !ok {
	// 			var status *model.Status
	// 			var err *model.AppError
	// 			if status, err = a.GetStatus(id); err != nil {
	// 				status = &model.Status{UserId: id, Status: model.STATUS_OFFLINE, Manual: false, LastActivityAt: 0, ActiveChannel: ""}
	// 			}

	// 			if ShouldSendPushNotification(profileMap[id], channelMemberNotifyPropsMap[id], false, status, post) {
	// 				a.sendPushNotification(
	// 					notification,
	// 					profileMap[id],
	// 					false,
	// 					false,
	// 					"",
	// 				)
	// 			} else {
	// 				// register that a notification was not sent
	// 				a.NotificationsLog().Warn("Notification not sent",
	// 					mlog.String("ackId", ""),
	// 					mlog.String("type", model.PUSH_TYPE_MESSAGE),
	// 					mlog.String("userId", id),
	// 					mlog.String("postId", post.Id),
	// 					mlog.String("status", model.PUSH_NOT_SENT),
	// 				)
	// 			}
	// 		}
	// 	}
	// }

	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_POSTED, "", post.ChannelId, "", nil)

	// // Note that PreparePostForClient should've already been called by this point
	// message.Add("post", post.ToJson())

	// message.Add("channel_type", channel.Type)
	// message.Add("channel_display_name", notification.GetChannelName(model.SHOW_USERNAME, ""))
	// message.Add("channel_name", channel.Name)
	// message.Add("sender_name", notification.GetSenderName(model.SHOW_USERNAME, *a.Config().ServiceSettings.EnablePostUsernameOverride))
	// message.Add("team_id", team.Id)
	// message.Add("set_online", setOnline)

	// if len(post.FileIds) != 0 && fchan != nil {
	// 	message.Add("otherFile", "true")

	// 	var infos []*model.FileInfo
	// 	if result := <-fchan; result.NErr != nil {
	// 		mlog.Warn("Unable to get fileInfo for push notifications.", mlog.String("post_id", post.Id), mlog.Err(result.NErr))
	// 	} else {
	// 		infos = result.Data.([]*model.FileInfo)
	// 	}

	// 	for _, info := range infos {
	// 		if info.IsImage() {
	// 			message.Add("image", "true")
	// 			break
	// 		}
	// 	}
	// }

	// if len(mentionedUsersList) != 0 {
	// 	message.Add("mentions", model.ArrayToJson(mentionedUsersList))
	// }

	// a.Publish(message)
	// // If this is a reply in a thread, notify participants
	// if *a.Config().ServiceSettings.CollapsedThreads != model.COLLAPSED_THREADS_DISABLED && post.RootId != "" {
	// 	thread, err := a.Srv().Store.Thread().Get(post.RootId)
	// 	if err != nil {
	// 		mlog.Error("Cannot get thread", mlog.String("id", post.RootId))
	// 		return nil, err
	// 	}
	// 	payload := thread.ToJson()
	// 	for _, uid := range thread.Participants {
	// 		sendEvent := *a.Config().ServiceSettings.CollapsedThreads == model.COLLAPSED_THREADS_DEFAULT_ON
	// 		// check if a participant has overridden collapsed threads settings
	// 		if preference, err := a.Srv().Store.Preference().Get(uid, model.PREFERENCE_CATEGORY_COLLAPSED_THREADS_SETTINGS, model.PREFERENCE_NAME_COLLAPSED_THREADS_ENABLED); err == nil {
	// 			sendEvent, _ = strconv.ParseBool(preference.Value)
	// 		}
	// 		if sendEvent {
	// 			message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_THREAD_UPDATED, "", "", uid, nil)
	// 			message.Add("thread", payload)
	// 			a.Publish(message)
	// 		}
	// 	}

	// }
	// return mentionedUsersList, nil
}
