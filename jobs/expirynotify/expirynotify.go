// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package expirynotify

import (
	"github.com/masterhung0112/hk_server/v5/app"
	tjobs "github.com/masterhung0112/hk_server/v5/jobs/interfaces"
)

type ExpiryNotifyJobInterfaceImpl struct {
	App *app.App
}

func init() {
	app.RegisterJobsExpiryNotifyJobInterface(func(s *app.Server) tjobs.ExpiryNotifyJobInterface {
		a := app.New(app.ServerConnector(s))
		return &ExpiryNotifyJobInterfaceImpl{a}
	})
}
