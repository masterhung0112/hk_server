// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package einterfaces

import "github.com/masterhung0112/hk_server/v5/model"

type LicenseInterface interface {
	CanStartTrial() (bool, error)
	GetPrevTrial() (*model.License, error)
}
