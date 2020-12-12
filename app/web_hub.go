package app

import (
	"github.com/masterhung0112/hk_server/model"
)

func (a *App) invalidateCacheForChannel(channel *model.Channel) {
	a.Srv().Store.Channel().InvalidateChannel(channel.Id)
	a.invalidateCacheForChannelByNameSkipClusterSend(channel.TeamId, channel.Name)

	if a.Cluster() != nil {
		nameMsg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_BY_NAME,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Props:    make(map[string]string),
		}

		nameMsg.Props["name"] = channel.Name
		if channel.TeamId == "" {
			nameMsg.Props["id"] = "dm"
		} else {
			nameMsg.Props["id"] = channel.TeamId
		}

		a.Cluster().SendClusterMessage(nameMsg)
	}
}

func (a *App) invalidateCacheForChannelMembers(channelId string) {
	a.Srv().Store.User().InvalidateProfilesInChannelCache(channelId)
	a.Srv().Store.Channel().InvalidateMemberCount(channelId)
	a.Srv().Store.Channel().InvalidateGuestCount(channelId)
}

func (a *App) invalidateCacheForChannelMembersNotifyProps(channelId string) {
	a.invalidateCacheForChannelMembersNotifyPropsSkipClusterSend(channelId)

	if a.Cluster() != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_MEMBERS_NOTIFY_PROPS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     channelId,
		}
		a.Cluster().SendClusterMessage(msg)
	}
}

func (a *App) invalidateCacheForChannelMembersNotifyPropsSkipClusterSend(channelId string) {
	a.Srv().Store.Channel().InvalidateCacheForChannelMembersNotifyProps(channelId)
}

func (a *App) invalidateCacheForChannelByNameSkipClusterSend(teamId, name string) {
	if teamId == "" {
		teamId = "dm"
	}

	a.Srv().Store.Channel().InvalidateChannelByName(teamId, name)
}

func (a *App) invalidateCacheForChannelPosts(channelId string) {
	a.Srv().Store.Channel().InvalidatePinnedPostCount(channelId)
	a.Srv().Store.Post().InvalidateLastPostTimeCache(channelId)
}

func (a *App) InvalidateCacheForUser(userId string) {
	a.invalidateCacheForUserSkipClusterSend(userId)

	a.Srv().Store.User().InvalidateProfilesInChannelCacheByUser(userId)
	a.Srv().Store.User().InvalidateProfileCacheForUser(userId)

	if a.Cluster() != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     userId,
		}
		a.Cluster().SendClusterMessage(msg)
	}
}

func (a *App) invalidateCacheForUserTeams(userId string) {
	a.InvalidateWebConnSessionCacheForUser(userId)
	a.Srv().Store.Team().InvalidateAllTeamIdsForUser(userId)

	if a.Cluster() != nil {
		msg := &model.ClusterMessage{
			Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER_TEAMS,
			SendType: model.CLUSTER_SEND_BEST_EFFORT,
			Data:     userId,
		}
		a.Cluster().SendClusterMessage(msg)
	}
}

func (a *App) invalidateCacheForUserSkipClusterSend(userId string) {
	a.Srv().Store.Channel().InvalidateAllChannelMembersForUser(userId)
	a.InvalidateWebConnSessionCacheForUser(userId)
}

func (a *App) invalidateCacheForWebhook(webhookId string) {
	// a.Srv().Store.Webhook().InvalidateWebhookCache(webhookId)
}

func (a *App) InvalidateWebConnSessionCacheForUser(userId string) {
	// hub := a.GetHubForUserId(userId)
	// if hub != nil {
	// 	hub.InvalidateUser(userId)
	// }
}
