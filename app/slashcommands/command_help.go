// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	goi18n "github.com/mattermost/go-i18n/i18n"
)

type HelpProvider struct {
}

const (
	CMD_HELP = "help"
)

func init() {
	app.RegisterCommandProvider(&HelpProvider{})
}

func (h *HelpProvider) GetTrigger() string {
	return CMD_HELP
}

func (h *HelpProvider) GetCommand(a *app.App, T goi18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CMD_HELP,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_help.desc"),
		DisplayName:      T("api.command_help.name"),
	}
}

func (h *HelpProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	helpLink := *a.Config().SupportSettings.HelpLink

	if helpLink == "" {
		helpLink = model.SUPPORT_SETTINGS_DEFAULT_HELP_LINK
	}

	return &model.CommandResponse{GotoLocation: helpLink}
}
