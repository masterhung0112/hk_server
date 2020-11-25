package app

import (
	"encoding/json"
	"errors"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/utils"
	"net/http"
)

// deduplicateCreatePost attempts to make posting idempotent within a caching window.
func (a *App) deduplicateCreatePost(post *model.Post) (foundPost *model.Post, err *model.AppError) {
	// We rely on the client sending the pending post id across "duplicate" requests. If there
	// isn't one, we can't deduplicate, so allow creation normally.
	if post.PendingPostId == "" {
		return nil, nil
	}

	const unknownPostId = ""

	// Query the cache atomically for the given pending post id, saving a record if
	// it hasn't previously been seen.
	//TODO: Open
	var postId string
	// nErr := a.Srv().seenPendingPostIdsCache.Get(post.PendingPostId, &postId)
	// if nErr == cache.ErrKeyNotFound {
	// 	a.Srv().seenPendingPostIdsCache.SetWithExpiry(post.PendingPostId, unknownPostId, PENDING_POST_IDS_CACHE_TTL)
	return nil, nil
	// }

	// if nErr != nil {
	// 	return nil, model.NewAppError("errorGetPostId", "api.post.error_get_post_id.pending", nil, "", http.StatusInternalServerError)
	// }

	// If another thread saved the cache record, but hasn't yet updated it with the actual post
	// id (because it's still saving), notify the client with an error. Ideally, we'd wait
	// for the other thread, but coordinating that adds complexity to the happy path.
	if postId == unknownPostId {
		return nil, model.NewAppError("deduplicateCreatePost", "api.post.deduplicate_create_post.pending", nil, "", http.StatusInternalServerError)
	}

	// If the other thread finished creating the post, return the created post back to the
	// client, making the API call feel idempotent.
	actualPost, err := a.GetSinglePost(postId)
	if err != nil {
		return nil, model.NewAppError("deduplicateCreatePost", "api.post.deduplicate_create_post.failed_to_get", nil, err.Error(), http.StatusInternalServerError)
	}

	mlog.Debug("Deduplicated create post", mlog.String("post_id", actualPost.Id), mlog.String("pending_post_id", post.PendingPostId))

	return actualPost, nil
}

func (a *App) CreatePost(post *model.Post, channel *model.Channel, triggerWebhooks, setOnline bool) (savedPost *model.Post, err *model.AppError) {
	foundPost, err := a.deduplicateCreatePost(post)
	if err != nil {
		return nil, err
	}
	if foundPost != nil {
		return foundPost, nil
	}

	// If we get this far, we've recorded the client-provided pending post id to the cache.
	// Remove it if we fail below, allowing a proper retry by the client.
	defer func() {
		if post.PendingPostId == "" {
			return
		}

		if err != nil {
			//TODO: Open
			// a.Srv().seenPendingPostIdsCache.Remove(post.PendingPostId)
			return
		}

		//TODO: Open
		// a.Srv().seenPendingPostIdsCache.SetWithExpiry(post.PendingPostId, savedPost.Id, PENDING_POST_IDS_CACHE_TTL)
	}()

	post.SanitizeProps()

	var pchan chan store.StoreResult
	if len(post.RootId) > 0 {
		pchan = make(chan store.StoreResult, 1)
		go func() {
			r, pErr := a.Srv().Store.Post().Get(post.RootId, false)
			pchan <- store.StoreResult{Data: r, NErr: pErr}
			close(pchan)
		}()
	}

	user, nErr := a.Srv().Store.User().Get(post.UserId)
	if nErr != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(nErr, &nfErr):
			return nil, model.NewAppError("CreatePost", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("CreatePost", "app.user.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if user.IsBot {
		post.AddProp("from_bot", "true")
	}

	if a.Srv().License() != nil && *a.Config().TeamSettings.ExperimentalTownSquareIsReadOnly &&
		!post.IsSystemMessage() &&
		channel.Name == model.DEFAULT_CHANNEL &&
		!a.RolesGrantPermission(user.GetRoles(), model.PERMISSION_MANAGE_SYSTEM.Id) {
		return nil, model.NewAppError("createPost", "api.post.create_post.town_square_read_only", nil, "", http.StatusForbidden)
	}

	var ephemeralPost *model.Post
	if post.Type == "" && !a.HasPermissionToChannel(user.Id, channel.Id, model.PERMISSION_USE_CHANNEL_MENTIONS) {
		mention := post.DisableMentionHighlights()
		if mention != "" {
			T := utils.GetUserTranslations(user.Locale)
			ephemeralPost = &model.Post{
				UserId:    user.Id,
				RootId:    post.RootId,
				ParentId:  post.ParentId,
				ChannelId: channel.Id,
				Message:   T("model.post.channel_notifications_disabled_in_channel.message", model.StringInterface{"ChannelName": channel.Name, "Mention": mention}),
				Props:     model.StringInterface{model.POST_PROPS_MENTION_HIGHLIGHT_DISABLED: true},
			}
		}
	}

	// Verify the parent/child relationships are correct
	var parentPostList *model.PostList
	if pchan != nil {
		result := <-pchan
		if result.NErr != nil {
			return nil, model.NewAppError("createPost", "api.post.create_post.root_id.app_error", nil, "", http.StatusBadRequest)
		}
		parentPostList = result.Data.(*model.PostList)
		if len(parentPostList.Posts) == 0 || !parentPostList.IsChannelId(post.ChannelId) {
			return nil, model.NewAppError("createPost", "api.post.create_post.channel_root_id.app_error", nil, "", http.StatusInternalServerError)
		}

		rootPost := parentPostList.Posts[post.RootId]
		if len(rootPost.RootId) > 0 {
			return nil, model.NewAppError("createPost", "api.post.create_post.root_id.app_error", nil, "", http.StatusBadRequest)
		}

		if post.ParentId == "" {
			post.ParentId = post.RootId
		}

		if post.RootId != post.ParentId {
			parent := parentPostList.Posts[post.ParentId]
			if parent == nil {
				return nil, model.NewAppError("createPost", "api.post.create_post.parent_id.app_error", nil, "", http.StatusInternalServerError)
			}
		}
	}

	post.Hashtags, _ = model.ParseHashtags(post.Message)

	if err = a.FillInPostProps(post, channel); err != nil {
		return nil, err
	}

	// Temporary fix so old plugins don't clobber new fields in SlackAttachment struct, see MM-13088
	if attachments, ok := post.GetProp("attachments").([]*model.SlackAttachment); ok {
		jsonAttachments, err := json.Marshal(attachments)
		if err == nil {
			attachmentsInterface := []interface{}{}
			err = json.Unmarshal(jsonAttachments, &attachmentsInterface)
			post.AddProp("attachments", attachmentsInterface)
		}
		if err != nil {
			mlog.Error("Could not convert post attachments to map interface.", mlog.Err(err))
		}
	}

	//TODO: Open
	// if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
	// 	var rejectionError *model.AppError
	// 	pluginContext := a.PluginContext()
	// 	pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	// 		replacementPost, rejectionReason := hooks.MessageWillBePosted(pluginContext, post)
	// 		if rejectionReason != "" {
	// 			id := "Post rejected by plugin. " + rejectionReason
	// 			if rejectionReason == plugin.DismissPostError {
	// 				id = plugin.DismissPostError
	// 			}
	// 			rejectionError = model.NewAppError("createPost", id, nil, "", http.StatusBadRequest)
	// 			return false
	// 		}
	// 		if replacementPost != nil {
	// 			post = replacementPost
	// 		}

	// 		return true
	// 	}, plugin.MessageWillBePostedId)

	// 	if rejectionError != nil {
	// 		return nil, rejectionError
	// 	}
	// }

	rpost, nErr := a.Srv().Store.Post().Save(post)
	if nErr != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &invErr):
			return nil, model.NewAppError("CreatePost", "app.post.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("CreatePost", "app.post.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	// Update the mapping from pending post id to the actual post id, for any clients that
	// might be duplicating requests.
	//TODO: Open
	// a.Srv().seenPendingPostIdsCache.SetWithExpiry(post.PendingPostId, rpost.Id, PENDING_POST_IDS_CACHE_TTL)

	// We make a copy of the post for the plugin hook to avoid a race condition.
	// rPostCopy := rpost.Clone()
	//TODO: Open
	// if pluginsEnvironment := a.GetPluginsEnvironment(); pluginsEnvironment != nil {
	// 	a.Srv().Go(func() {
	// 		pluginContext := a.PluginContext()
	// 		pluginsEnvironment.RunMultiPluginHook(func(hooks plugin.Hooks) bool {
	// 			hooks.MessageHasBeenPosted(pluginContext, rPostCopy)
	// 			return true
	// 		}, plugin.MessageHasBeenPostedId)
	// 	})
	// }

	//TODO: Open
	// if a.Metrics() != nil {
	// 	a.Metrics().IncrementPostCreate()
	// }

	if len(post.FileIds) > 0 {
		if err = a.attachFilesToPost(post); err != nil {
			mlog.Error("Encountered error attaching files to post", mlog.String("post_id", post.Id), mlog.Any("file_ids", post.FileIds), mlog.Err(err))
		}

		//TODO: Open
		// if a.Metrics() != nil {
		// 	a.Metrics().IncrementPostFileAttachment(len(post.FileIds))
		// }
	}

	// Normally, we would let the API layer call PreparePostForClient, but we do it here since it also needs
	// to be done when we send the post over the websocket in handlePostEvents
	rpost = a.PreparePostForClient(rpost, true, false)

	if err := a.handlePostEvents(rpost, user, channel, triggerWebhooks, parentPostList, setOnline); err != nil {
		mlog.Error("Failed to handle post events", mlog.Err(err))
	}

	// Send any ephemeral posts after the post is created to ensure it shows up after the latest post created
	if ephemeralPost != nil {
		a.SendEphemeralPost(post.UserId, ephemeralPost)
	}

	return rpost, nil
}

func (a *App) GetSinglePost(postId string) (*model.Post, *model.AppError) {
	post, err := a.Srv().Store.Post().GetSingle(postId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetSinglePost", "app.post.get.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetSinglePost", "app.post.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return post, nil
}

func (a *App) attachFilesToPost(post *model.Post) *model.AppError {
	//TODO: Open
	// var attachedIds []string
	// for _, fileId := range post.FileIds {
	// 	err := a.Srv().Store.FileInfo().AttachToPost(fileId, post.Id, post.UserId)
	// 	if err != nil {
	// 		mlog.Warn("Failed to attach file to post", mlog.String("file_id", fileId), mlog.String("post_id", post.Id), mlog.Err(err))
	// 		continue
	// 	}

	// 	attachedIds = append(attachedIds, fileId)
	// }

	// if len(post.FileIds) != len(attachedIds) {
	// 	// We couldn't attach all files to the post, so ensure that post.FileIds reflects what was actually attached
	// 	post.FileIds = attachedIds

	// 	if _, err := a.Srv().Store.Post().Overwrite(post); err != nil {
	// 		return model.NewAppError("attachFilesToPost", "app.post.overwrite.app_error", nil, err.Error(), http.StatusInternalServerError)
	// 	}
	// }

	return nil
}

// FillInPostProps should be invoked before saving posts to fill in properties such as
// channel_mentions.
//
// If channel is nil, FillInPostProps will look up the channel corresponding to the post.
func (a *App) FillInPostProps(post *model.Post, channel *model.Channel) *model.AppError {
	channelMentions := post.ChannelMentions()
	channelMentionsProp := make(map[string]interface{})

	if len(channelMentions) > 0 {
		if channel == nil {
			postChannel, err := a.Srv().Store.Channel().GetForPost(post.Id)
			if err != nil {
				return model.NewAppError("FillInPostProps", "api.context.invalid_param.app_error", map[string]interface{}{"Name": "post.channel_id"}, err.Error(), http.StatusBadRequest)
			}
			channel = postChannel
		}

		mentionedChannels, err := a.GetChannelsByNames(channelMentions, channel.TeamId)
		if err != nil {
			return err
		}

		for _, mentioned := range mentionedChannels {
			if mentioned.Type == model.CHANNEL_OPEN {
				team, err := a.Srv().Store.Team().Get(mentioned.TeamId)
				if err != nil {
					mlog.Error("Failed to get team of the channel mention", mlog.String("team_id", channel.TeamId), mlog.String("channel_id", channel.Id), mlog.Err(err))
					continue
				}
				channelMentionsProp[mentioned.Name] = map[string]interface{}{
					"display_name": mentioned.DisplayName,
					"team_name":    team.Name,
				}
			}
		}
	}

	if len(channelMentionsProp) > 0 {
		post.AddProp("channel_mentions", channelMentionsProp)
	} else if post.GetProps() != nil {
		post.DelProp("channel_mentions")
	}

	matched := model.AT_MENTION_PATTEN.MatchString(post.Message)
	if a.Srv().License() != nil && *a.Srv().License().Features.LDAPGroups && matched && !a.HasPermissionToChannel(post.UserId, post.ChannelId, model.PERMISSION_USE_GROUP_MENTIONS) {
		post.AddProp(model.POST_PROPS_GROUP_HIGHLIGHT_DISABLED, true)
	}

	return nil
}

func (a *App) handlePostEvents(post *model.Post, user *model.User, channel *model.Channel, triggerWebhooks bool, parentPostList *model.PostList, setOnline bool) error {
	var team *model.Team
	if len(channel.TeamId) > 0 {
		t, err := a.Srv().Store.Team().Get(channel.TeamId)
		if err != nil {
			return err
		}
		team = t
	} else {
		// Blank team for DMs
		team = &model.Team{}
	}

	a.invalidateCacheForChannel(channel)
	a.invalidateCacheForChannelPosts(channel.Id)

	if _, err := a.SendNotifications(post, team, channel, user, parentPostList, setOnline); err != nil {
		return err
	}

	//TODO: Open
	// if *a.Config().ServiceSettings.ThreadAutoFollow && post.RootId != "" {
	// 	if err := a.Srv().Store.Thread().CreateMembershipIfNeeded(post.UserId, post.RootId, true); err != nil {
	// 		return err
	// 	}
	// }

	if post.Type != model.POST_AUTO_RESPONDER { // don't respond to an auto-responder
		a.Srv().Go(func() {
			_, err := a.SendAutoResponseIfNecessary(channel, user, post)
			if err != nil {
				mlog.Error("Failed to send auto response", mlog.String("user_id", user.Id), mlog.String("post_id", post.Id), mlog.Err(err))
			}
		})
	}

	if triggerWebhooks {
		a.Srv().Go(func() {
			if err := a.handleWebhookEvents(post, team, channel, user); err != nil {
				mlog.Error(err.Error())
			}
		})
	}

	return nil
}

func (a *App) SendEphemeralPost(userId string, post *model.Post) *model.Post {
	post.Type = model.POST_EPHEMERAL

	// fill in fields which haven't been specified which have sensible defaults
	if post.Id == "" {
		post.Id = model.NewId()
	}
	if post.CreateAt == 0 {
		post.CreateAt = model.GetMillis()
	}
	if post.GetProps() == nil {
		post.SetProps(make(model.StringInterface))
	}

	post.GenerateActionIds()
	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_EPHEMERAL_MESSAGE, "", post.ChannelId, userId, nil)
	// post = a.PreparePostForClient(post, true, false)
	// post = model.AddPostActionCookies(post, a.PostActionCookieSecret())
	// message.Add("post", post.ToJson())
	// a.Publish(message)

	return post
}
