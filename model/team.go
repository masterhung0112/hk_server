package model

import ()
import "strings"
import "net/http"
import "unicode/utf8"
import "io"
import "encoding/json"

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

func TeamFromJson(data io.Reader) *Team {
	var o *Team
	json.NewDecoder(data).Decode(&o)
	return o
}

func (o *Team) ToJson() string {
	b, _ := json.Marshal(o)
	return string(b)
}

func (o *Team) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
	o.UpdateAt = o.CreateAt

	o.Name = SanitizeUnicode(o.Name)
	o.DisplayName = SanitizeUnicode(o.DisplayName)
	o.Description = SanitizeUnicode(o.Description)
	o.CompanyName = SanitizeUnicode(o.CompanyName)

	if len(o.InviteId) == 0 {
		o.InviteId = NewId()
	}
}

func (o *Team) IsValid() *AppError {

	if !IsValidId(o.Id) {
		return NewAppError("Team.IsValid", "model.team.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	if o.CreateAt == 0 {
		return NewAppError("Team.IsValid", "model.team.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if o.UpdateAt == 0 {
		return NewAppError("Team.IsValid", "model.team.is_valid.update_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Email) > TEAM_EMAIL_MAX_LENGTH {
		return NewAppError("Team.IsValid", "model.team.is_valid.email.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Email) > 0 && !IsValidEmail(o.Email) {
		return NewAppError("Team.IsValid", "model.team.is_valid.email.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if utf8.RuneCountInString(o.DisplayName) == 0 || utf8.RuneCountInString(o.DisplayName) > TEAM_DISPLAY_NAME_MAX_RUNES {
		return NewAppError("Team.IsValid", "model.team.is_valid.name.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Name) > TEAM_NAME_MAX_LENGTH {
		return NewAppError("Team.IsValid", "model.team.is_valid.url.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.Description) > TEAM_DESCRIPTION_MAX_LENGTH {
		return NewAppError("Team.IsValid", "model.team.is_valid.description.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.InviteId) == 0 {
		return NewAppError("Team.IsValid", "model.team.is_valid.invite_id.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if IsReservedTeamName(o.Name) {
		return NewAppError("Team.IsValid", "model.team.is_valid.reserved.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if !IsValidTeamName(o.Name) {
		return NewAppError("Team.IsValid", "model.team.is_valid.characters.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if !(o.Type == TEAM_OPEN || o.Type == TEAM_INVITE) {
		return NewAppError("Team.IsValid", "model.team.is_valid.type.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.CompanyName) > TEAM_COMPANY_NAME_MAX_LENGTH {
		return NewAppError("Team.IsValid", "model.team.is_valid.company.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	if len(o.AllowedDomains) > TEAM_ALLOWED_DOMAINS_MAX_LENGTH {
		return NewAppError("Team.IsValid", "model.team.is_valid.domains.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}

func IsReservedTeamName(s string) bool {
	s = strings.ToLower(s)

	for _, value := range reservedName {
		if strings.Index(s, value) == 0 {
			return true
		}
	}

	return false
}

func IsValidTeamName(s string) bool {
	if !IsValidAlphaNum(s) {
		return false
	}

	if len(s) < TEAM_NAME_MIN_LENGTH {
		return false
	}

	return true
}
