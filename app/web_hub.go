package app

import (
	"github.com/masterhung0112/hk_server/model"
)

func (a *App) invalidateCacheForChannel(channel *model.Channel) {
	a.Srv().Store.Channel().InvalidateChannel(channel.Id)
	a.invalidateCacheForChannelByNameSkipClusterSend(channel.TeamId, channel.Name)

	//TODO: Open
	// if a.Cluster() != nil {
	// 	nameMsg := &model.ClusterMessage{
	// 		Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_CHANNEL_BY_NAME,
	// 		SendType: model.CLUSTER_SEND_BEST_EFFORT,
	// 		Props:    make(map[string]string),
	// 	}

	// 	nameMsg.Props["name"] = channel.Name
	// 	if channel.TeamId == "" {
	// 		nameMsg.Props["id"] = "dm"
	// 	} else {
	// 		nameMsg.Props["id"] = channel.TeamId
	// 	}

	// 	a.Cluster().SendClusterMessage(nameMsg)
	// }
}

func (a *App) invalidateCacheForChannelByNameSkipClusterSend(teamId, name string) {
	if teamId == "" {
		teamId = "dm"
	}

	a.Srv().Store.Channel().InvalidateChannelByName(teamId, name)
}

func (a *App) InvalidateCacheForUser(userId string) {
	//TODO: Open
	// a.invalidateCacheForUserSkipClusterSend(userId)

	// a.Srv().Store.User().InvalidateProfilesInChannelCacheByUser(userId)
	// a.Srv().Store.User().InvalidateProfileCacheForUser(userId)

	// if a.Cluster() != nil {
	// 	msg := &model.ClusterMessage{
	// 		Event:    model.CLUSTER_EVENT_INVALIDATE_CACHE_FOR_USER,
	// 		SendType: model.CLUSTER_SEND_BEST_EFFORT,
	// 		Data:     userId,
	// 	}
	// 	a.Cluster().SendClusterMessage(msg)
	// }
}
