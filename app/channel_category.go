package app

import (
	"github.com/masterhung0112/go_server/model"
	"net/http"
)

func (a *App) createInitialSidebarCategories(userId, teamId string) *model.AppError {
	nErr := a.Srv().Store.Channel().CreateInitialSidebarCategories(userId, teamId)

	if nErr != nil {
		return model.NewAppError("createInitialSidebarCategories", "app.channel.create_initial_sidebar_categories.internal_error", nil, nErr.Error(), http.StatusInternalServerError)
	}

	return nil
}

func (a *App) GetSidebarCategory(categoryId string) (*model.SidebarCategoryWithChannels, *model.AppError) {
	return a.Srv().Store.Channel().GetSidebarCategory(categoryId)
}
