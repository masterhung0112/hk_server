// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package main

import (
	"github.com/masterhung0112/hk_server/app/plugin_api_tests"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/plugin"
)

type MyPlugin struct {
	plugin.MattermostPlugin
	configuration plugin_api_tests.BasicConfig
}

func (p *MyPlugin) OnConfigurationChange() error {
	if err := p.API.LoadPluginConfiguration(&p.configuration); err != nil {
		return err
	}
	return nil
}

func (p *MyPlugin) MessageWillBePosted(c *plugin.Context, post *model.Post) (*model.Post, string) {

	// check existing user first
	data, err := p.API.GetProfileImage(p.configuration.BasicUserId)
	if err != nil {
		return nil, err.Error()
	}
	if plugin_api_tests.IsEmpty(data) {
		return nil, "GetProfileImage return empty"
	}

	// then unknown user
	data, err = p.API.GetProfileImage(model.NewId())
	if err == nil || data != nil {
		return nil, "GetProfileImage should've returned an error"
	}
	return nil, "OK"
}

func main() {
	plugin.ClientMain(&MyPlugin{})
}