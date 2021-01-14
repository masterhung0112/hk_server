// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package interfaces

import (
	"github.com/masterhung0112/hk_server/model"
)

type PluginsJobInterface interface {
	MakeWorker() model.Worker
	MakeScheduler() model.Scheduler
}
