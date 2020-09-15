package model

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"unicode/utf8"
)

const (
	CHANNEL_OPEN                   = "O"
	CHANNEL_PRIVATE                = "P"
	CHANNEL_DIRECT                 = "D"
	CHANNEL_GROUP                  = "G"
	CHANNEL_GROUP_MAX_USERS        = 8
	CHANNEL_GROUP_MIN_USERS        = 3
	DEFAULT_CHANNEL                = "town-square"
	CHANNEL_DISPLAY_NAME_MAX_RUNES = 64
	CHANNEL_NAME_MIN_LENGTH        = 2
	CHANNEL_NAME_MAX_LENGTH        = 64
	CHANNEL_HEADER_MAX_RUNES       = 1024
	CHANNEL_PURPOSE_MAX_RUNES      = 250
	CHANNEL_CACHE_SIZE             = 25000

	CHANNEL_SORT_BY_USERNAME = "username"
	CHANNEL_SORT_BY_STATUS   = "status"
)

type Channel struct {
	Id               string                 `json:"id"`
	CreateAt         int64                  `json:"create_at"`
	UpdateAt         int64                  `json:"update_at"`
	DeleteAt         int64                  `json:"delete_at"`
	TeamId           string                 `json:"team_id"`
	Type             string                 `json:"type"`
	DisplayName      string                 `json:"display_name"`
	Name             string                 `json:"name"`
	Header           string                 `json:"header"`
	Purpose          string                 `json:"purpose"`
	LastPostAt       int64                  `json:"last_post_at"`
	TotalMsgCount    int64                  `json:"total_msg_count"`
	ExtraUpdateAt    int64                  `json:"extra_update_at"`
	CreatorId        string                 `json:"creator_id"`
	SchemeId         *string                `json:"scheme_id"`
	Props            map[string]interface{} `json:"props" db:"-"`
	GroupConstrained *bool                  `json:"group_constrained"`
}

type ChannelWithTeamData struct {
	Channel
	TeamDisplayName string `json:"team_display_name"`
	TeamName        string `json:"team_name"`
	TeamUpdateAt    int64  `json:"team_update_at"`
}

type ChannelsWithCount struct {
	Channels   *ChannelListWithTeamData `json:"channels"`
	TotalCount int64                    `json:"total_count"`
}

type ChannelPatch struct {
	DisplayName      *string `json:"display_name"`
	Name             *string `json:"name"`
	Header           *string `json:"header"`
	Purpose          *string `json:"purpose"`
	GroupConstrained *bool   `json:"group_constrained"`
}

type ChannelForExport struct {
	Channel
	TeamName   string
	SchemeName *string
}

type DirectChannelForExport struct {
	Channel
	Members *[]string
}

type ChannelModeration struct {
	Name  string                 `json:"name"`
	Roles *ChannelModeratedRoles `json:"roles"`
}

type ChannelModeratedRoles struct {
	Guests  *ChannelModeratedRole `json:"guests"`
	Members *ChannelModeratedRole `json:"members"`
}

type ChannelModeratedRole struct {
	Value   bool `json:"value"`
	Enabled bool `json:"enabled"`
}

type ChannelModerationPatch struct {
	Name  *string                     `json:"name"`
	Roles *ChannelModeratedRolesPatch `json:"roles"`
}

type ChannelModeratedRolesPatch struct {
	Guests  *bool `json:"guests"`
	Members *bool `json:"members"`
}

// ChannelSearchOpts contains options for searching channels.
//
// NotAssociatedToGroup will exclude channels that have associated, active GroupChannels records.
// ExcludeDefaultChannels will exclude the configured default channels (ex 'town-square' and 'off-topic').
// IncludeDeleted will include channel records where DeleteAt != 0.
// ExcludeChannelNames will exclude channels from the results by name.
// Paginate whether to paginate the results.
// Page page requested, if results are paginated.
// PerPage number of results per page, if paginated.
//
type ChannelSearchOpts struct {
	NotAssociatedToGroup    string
	ExcludeDefaultChannels  bool
	IncludeDeleted          bool
	Deleted                 bool
	ExcludeChannelNames     []string
	TeamIds                 []string
	GroupConstrained        bool
	ExcludeGroupConstrained bool
	Public                  bool
	Private                 bool
	Page                    *int
	PerPage                 *int
}

type ChannelMemberCountByGroup struct {
	GroupId                     string `db:"-" json:"group_id"`
	ChannelMemberCount          int64  `db:"-" json:"channel_member_count"`
	ChannelMemberTimezonesCount int64  `db:"-" json:"channel_member_timezones_count"`
}

func (o *Channel) IsValid() *AppError {
	if !IsValidId(o.Id) {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.DisplayName) > CHANNEL_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.display_name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if !IsValidChannelIdentifier(o.Name) {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.2_or_more.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if !(o.Type == CHANNEL_OPEN || o.Type == CHANNEL_PRIVATE || o.Type == CHANNEL_DIRECT || o.Type == CHANNEL_GROUP) {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.type.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Header) > CHANNEL_HEADER_MAX_RUNES {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.header.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.Purpose) > CHANNEL_PURPOSE_MAX_RUNES {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.purpose.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.CreatorId) > 26 {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.creator_id.app_error", nil, "", http.StatusBadRequest)
	}

	userIds := strings.Split(o.Name, "__")
	if o.Type != CHANNEL_DIRECT && len(userIds) == 2 && IsValidId(userIds[0]) && IsValidId(userIds[1]) {
		return NewAppError("Channel.IsValid", "model.channel.is_valid.name.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}

func (o *Channel) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.Name = SanitizeUnicode(o.Name)
	o.DisplayName = SanitizeUnicode(o.DisplayName)

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt
	o.ExtraUpdateAt = 0
}

func (o *Channel) IsGroupConstrained() bool {
	return o.GroupConstrained != nil && *o.GroupConstrained
}

func (o *Channel) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func ChannelFromJson(data io.Reader) *Channel {
	var o *Channel
	json.NewDecoder(data).Decode(&o)
	return o
}
