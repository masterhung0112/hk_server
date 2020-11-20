package model

import (
	"sync"
)

type Post struct {
	Id         string `json:"id"`
	CreateAt   int64  `json:"create_at"`
	UpdateAt   int64  `json:"update_at"`
	EditAt     int64  `json:"edit_at"`
	DeleteAt   int64  `json:"delete_at"`
	IsPinned   bool   `json:"is_pinned"`
	UserId     string `json:"user_id"`
	ChannelId  string `json:"channel_id"`
	RootId     string `json:"root_id"`
	ParentId   string `json:"parent_id"`
	OriginalId string `json:"original_id"`

	Message string `json:"message"`
	// MessageSource will contain the message as submitted by the user if Message has been modified
	// by Mattermost for presentation (e.g if an image proxy is being used). It should be used to
	// populate edit boxes if present.
	MessageSource string `json:"message_source,omitempty" db:"-"`

	Type          string          `json:"type"`
	propsMu       sync.RWMutex    `db:"-"`       // Unexported mutex used to guard Post.Props.
	Props         StringInterface `json:"props"` // Deprecated: use GetProps()
	Hashtags      string          `json:"hashtags"`
	Filenames     StringArray     `json:"filenames,omitempty"` // Deprecated, do not use this field any more
	FileIds       StringArray     `json:"file_ids,omitempty"`
	PendingPostId string          `json:"pending_post_id" db:"-"`
	HasReactions  bool            `json:"has_reactions,omitempty"`

	// Transient data populated before sending a post to the client
	ReplyCount int64         `json:"reply_count" db:"-"`
	Metadata   *PostMetadata `json:"metadata,omitempty" db:"-"`
}

type GetPostsSinceOptions struct {
	ChannelId        string
	Time             int64
	SkipFetchThreads bool
}

type GetPostsOptions struct {
	ChannelId        string
	PostId           string
	Page             int
	PerPage          int
	SkipFetchThreads bool
}
