// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	goi18n "github.com/mattermost/go-i18n/i18n"
)

type SettingsProvider struct {
}

const (
	CMD_SETTINGS = "settings"
)

func init() {
	app.RegisterCommandProvider(&SettingsProvider{})
}

func (settings *SettingsProvider) GetTrigger() string {
	return CMD_SETTINGS
}

func (settings *SettingsProvider) GetCommand(a *app.App, T goi18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CMD_SETTINGS,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_settings.desc"),
		AutoCompleteHint: "",
		DisplayName:      T("api.command_settings.name"),
	}
}

func (settings *SettingsProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	// This command is handled client-side and shouldn't hit the server.
	return &model.CommandResponse{
		Text:         args.T("api.command_settings.unsupported.app_error"),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}
}
