package model

import (
	"strings"
)

const (
  SESSION_CACHE_SIZE                = 35000
)

// Session contains the user session details.
// This struct's serializer methods are auto-generated. If a new field is added/removed,
// please run make gen-serialized.
type Session struct {
  Id             string        `json:"id"`
  UserId         string        `json:"user_id"`
  Roles          string        `json:"roles"`
  TeamMembers    []*TeamMember `json:"team_members" db:"-"`
  Local          bool          `json:"local" db:"-"`
}

// Returns true if the session is unrestricted, which should grant it
// with all permissions. This is used for local mode sessions
func (me *Session) IsUnrestricted() bool {
	return me.Local
}

func (me *Session) GetUserRoles() []string {
	return strings.Fields(me.Roles)
}

func (me *Session) GetTeamByTeamId(teamId string) *TeamMember {
	for _, team := range me.TeamMembers {
		if team.TeamId == teamId {
			return team
		}
	}

	return nil
}