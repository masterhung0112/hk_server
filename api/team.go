package api

import (
	"github.com/masterhung0112/go_server/model"
	"net/http"
	"strings"
)

func (api *API) InitTeam() {
	api.BaseRoutes.Teams.Handle("", api.ApiSessionRequired(createTeam)).Methods("POST")
}

func createTeam(c *Context, w http.ResponseWriter, r *http.Request) {
	team := model.TeamFromJson(r.Body)
	if team == nil {
		c.SetInvalidParam("team")
		return
	}
	team.Email = strings.ToLower(team.Email)

	// auditRec := c.MakeAuditRecord("createTeam", audit.Fail)
	// defer c.LogAuditRec(auditRec)
	// auditRec.AddMeta("team", team)

	if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_CREATE_TEAM) {
		c.Err = model.NewAppError("createTeam", "api.team.is_team_creation_allowed.disabled.app_error", nil, "", http.StatusForbidden)
		return
	}

	rteam, err := c.App.CreateTeamWithUser(team, c.App.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	// Don't sanitize the team here since the user will be a team admin and their session won't reflect that yet

	//TODO: Open
	// auditRec.Success()
	// auditRec.AddMeta("team", team) // overwrite meta

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(rteam.ToJson()))
}
