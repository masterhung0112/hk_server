// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/shared/i18n"
)

type ShrugProvider struct {
}

const (
	CmdShrug = "shrug"
)

func init() {
	app.RegisterCommandProvider(&ShrugProvider{})
}

func (*ShrugProvider) GetTrigger() string {
	return CmdShrug
}

func (*ShrugProvider) GetCommand(a *app.App, T i18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CmdShrug,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_shrug.desc"),
		AutoCompleteHint: T("api.command_shrug.hint"),
		DisplayName:      T("api.command_shrug.name"),
	}
}

func (*ShrugProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	rmsg := `¯\\\_(ツ)\_/¯`
	if message != "" {
		rmsg = message + " " + rmsg
	}

	return &model.CommandResponse{ResponseType: model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, Text: rmsg}
}
