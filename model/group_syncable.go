package model

import (
	"net/http"
)

type GroupSyncableType string

const (
	GroupSyncableTypeTeam    GroupSyncableType = "Team"
	GroupSyncableTypeChannel GroupSyncableType = "Channel"
)

func (gst GroupSyncableType) String() string {
	return string(gst)
}

type GroupSyncable struct {
	GroupId string `json:"group_id"`

	// SyncableId represents the Id of the model that is being synced with the group, for example a ChannelId or
	// TeamId.
	SyncableId string `db:"-" json:"-"`

	AutoAdd     bool              `json:"auto_add"`
	SchemeAdmin bool              `json:"scheme_admin"`
	CreateAt    int64             `json:"create_at"`
	DeleteAt    int64             `json:"delete_at"`
	UpdateAt    int64             `json:"update_at"`
	Type        GroupSyncableType `db:"-" json:"-"`

	// Values joined in from the associated team and/or channel
	ChannelDisplayName string `db:"-" json:"-"`
	TeamDisplayName    string `db:"-" json:"-"`
	TeamType           string `db:"-" json:"-"`
	ChannelType        string `db:"-" json:"-"`
	TeamID             string `db:"-" json:"-"`
}

func (syncable *GroupSyncable) IsValid() *AppError {
	if !IsValidId(syncable.GroupId) {
		return NewAppError("GroupSyncable.SyncableIsValid", "model.group_syncable.group_id.app_error", nil, "", http.StatusBadRequest)
	}
	if !IsValidId(syncable.SyncableId) {
		return NewAppError("GroupSyncable.SyncableIsValid", "model.group_syncable.syncable_id.app_error", nil, "", http.StatusBadRequest)
	}
	return nil
}
