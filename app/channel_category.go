package app

import (
	"github.com/masterhung0112/go_server/model"
)

func (a *App) GetSidebarCategory(categoryId string) (*model.SidebarCategoryWithChannels, *model.AppError) {
	return a.Srv().Store.Channel().GetSidebarCategory(categoryId)
}
