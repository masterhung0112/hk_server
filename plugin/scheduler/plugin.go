// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package scheduler

import (
	"github.com/masterhung0112/hk_server/app"
	tjobs "github.com/masterhung0112/hk_server/jobs/interfaces"
)

type PluginsJobInterfaceImpl struct {
	App *app.App
}

func init() {
	app.RegisterJobsPluginsJobInterface(func(a *app.App) tjobs.PluginsJobInterface {
		return &PluginsJobInterfaceImpl{a}
	})
}
