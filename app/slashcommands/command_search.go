// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	goi18n "github.com/mattermost/go-i18n/i18n"
)

type SearchProvider struct {
}

const (
	CMD_SEARCH = "search"
)

func init() {
	app.RegisterCommandProvider(&SearchProvider{})
}

func (search *SearchProvider) GetTrigger() string {
	return CMD_SEARCH
}

func (search *SearchProvider) GetCommand(a *app.App, T goi18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CMD_SEARCH,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_search.desc"),
		AutoCompleteHint: T("api.command_search.hint"),
		DisplayName:      T("api.command_search.name"),
	}
}

func (search *SearchProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	// This command is handled client-side and shouldn't hit the server.
	return &model.CommandResponse{
		Text:         args.T("api.command_search.unsupported.app_error"),
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
	}
}
