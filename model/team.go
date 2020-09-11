package model

import ()

const (
	TEAM_OPEN                       = "O"
	TEAM_INVITE                     = "I"
	TEAM_ALLOWED_DOMAINS_MAX_LENGTH = 500
	TEAM_COMPANY_NAME_MAX_LENGTH    = 64
	TEAM_DESCRIPTION_MAX_LENGTH     = 255
	TEAM_DISPLAY_NAME_MAX_RUNES     = 64
	TEAM_EMAIL_MAX_LENGTH           = 128
	TEAM_NAME_MAX_LENGTH            = 64
	TEAM_NAME_MIN_LENGTH            = 2
)

type Team struct {
	Id                 string  `json:"id"`
	CreateAt           int64   `json:"create_at"`
	UpdateAt           int64   `json:"update_at"`
	DeleteAt           int64   `json:"delete_at"`
	DisplayName        string  `json:"display_name"`
	Name               string  `json:"name"`
	Description        string  `json:"description"`
	Email              string  `json:"email"`
	Type               string  `json:"type"`
	CompanyName        string  `json:"company_name"`
	AllowedDomains     string  `json:"allowed_domains"`
	InviteId           string  `json:"invite_id"`
	AllowOpenInvite    bool    `json:"allow_open_invite"`
	LastTeamIconUpdate int64   `json:"last_team_icon_update,omitempty"`
	SchemeId           *string `json:"scheme_id"`
	GroupConstrained   *bool   `json:"group_constrained"`
}

type TeamPatch struct {
	DisplayName      *string `json:"display_name"`
	Description      *string `json:"description"`
	CompanyName      *string `json:"company_name"`
	AllowedDomains   *string `json:"allowed_domains"`
	AllowOpenInvite  *bool   `json:"allow_open_invite"`
	GroupConstrained *bool   `json:"group_constrained"`
}

type TeamForExport struct {
	Team
	SchemeName *string
}

type Invites struct {
	Invites []map[string]string `json:"invites"`
}

type TeamsWithCount struct {
	Teams      []*Team `json:"teams"`
	TotalCount int64   `json:"total_count"`
}
