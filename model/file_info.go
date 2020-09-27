package model

import ()

const (
	FILEINFO_SORT_BY_CREATED = "CreateAt"
	FILEINFO_SORT_BY_SIZE    = "Size"
)

// GetFileInfosOptions contains options for getting FileInfos
type GetFileInfosOptions struct {
	// UserIds optionally limits the FileInfos to those created by the given users.
	UserIds []string `json:"user_ids"`
	// ChannelIds optionally limits the FileInfos to those created in the given channels.
	ChannelIds []string `json:"channel_ids"`
	// Since optionally limits FileInfos to those created at or after the given time, specified as Unix time in milliseconds.
	Since int64 `json:"since"`
	// IncludeDeleted if set includes deleted FileInfos.
	IncludeDeleted bool `json:"include_deleted"`
	// SortBy sorts the FileInfos by this field. The default is to sort by date created.
	SortBy string `json:"sort_by"`
	// SortDescending changes the sort direction to descending order when true.
	SortDescending bool `json:"sort_descending"`
}

type FileInfo struct {
	Id              string `json:"id"`
	CreatorId       string `json:"user_id"`
	PostId          string `json:"post_id,omitempty"`
	CreateAt        int64  `json:"create_at"`
	UpdateAt        int64  `json:"update_at"`
	DeleteAt        int64  `json:"delete_at"`
	Path            string `json:"-"` // not sent back to the client
	ThumbnailPath   string `json:"-"` // not sent back to the client
	PreviewPath     string `json:"-"` // not sent back to the client
	Name            string `json:"name"`
	Extension       string `json:"extension"`
	Size            int64  `json:"size"`
	MimeType        string `json:"mime_type"`
	Width           int    `json:"width,omitempty"`
	Height          int    `json:"height,omitempty"`
	HasPreviewImage bool   `json:"has_preview_image,omitempty"`
}
