package app

import (
	"github.com/masterhung0112/go_server/model"
)

func (a *App) GetTeamMember(teamId, userId string) (*model.TeamMember, *model.AppError) {
	return a.Srv().Store.Team().GetMember(teamId, userId)
}
