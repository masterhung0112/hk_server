package model

import (
	"strings"
)

const (
	CHANNEL_NOTIFY_DEFAULT              = "default"
	CHANNEL_NOTIFY_ALL                  = "all"
	CHANNEL_NOTIFY_MENTION              = "mention"
	CHANNEL_NOTIFY_NONE                 = "none"
	CHANNEL_MARK_UNREAD_ALL             = "all"
	CHANNEL_MARK_UNREAD_MENTION         = "mention"
	IGNORE_CHANNEL_MENTIONS_DEFAULT     = "default"
	IGNORE_CHANNEL_MENTIONS_OFF         = "off"
	IGNORE_CHANNEL_MENTIONS_ON          = "on"
	IGNORE_CHANNEL_MENTIONS_NOTIFY_PROP = "ignore_channel_mentions"
)

type ChannelUnread struct {
	TeamId       string    `json:"team_id"`
	ChannelId    string    `json:"channel_id"`
	MsgCount     int64     `json:"msg_count"`
	MentionCount int64     `json:"mention_count"`
	NotifyProps  StringMap `json:"-"`
}

type ChannelUnreadAt struct {
	TeamId       string    `json:"team_id"`
	UserId       string    `json:"user_id"`
	ChannelId    string    `json:"channel_id"`
	MsgCount     int64     `json:"msg_count"`
	MentionCount int64     `json:"mention_count"`
	LastViewedAt int64     `json:"last_viewed_at"`
	NotifyProps  StringMap `json:"-"`
}

type ChannelMember struct {
	ChannelId     string    `json:"channel_id"`
	UserId        string    `json:"user_id"`
	Roles         string    `json:"roles"`
	LastViewedAt  int64     `json:"last_viewed_at"`
	MsgCount      int64     `json:"msg_count"`
	MentionCount  int64     `json:"mention_count"`
	NotifyProps   StringMap `json:"notify_props"`
	LastUpdateAt  int64     `json:"last_update_at"`
	SchemeGuest   bool      `json:"scheme_guest"`
	SchemeUser    bool      `json:"scheme_user"`
	SchemeAdmin   bool      `json:"scheme_admin"`
	ExplicitRoles string    `json:"explicit_roles"`
}

type ChannelMembers []ChannelMember

type ChannelMemberForExport struct {
	ChannelMember
	ChannelName string
	Username    string
}

func (o *ChannelMember) GetRoles() []string {
	return strings.Fields(o.Roles)
}
