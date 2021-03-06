// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/masterhung0112/hk_server/v5/model"
)

func TestMeProviderDoCommand(t *testing.T) {
	th := setup(t)
	defer th.tearDown()

	mp := MeProvider{}

	msg := "hello"

	resp := mp.DoCommand(th.App, th.Context, &model.CommandArgs{}, msg)

	assert.Equal(t, model.COMMAND_RESPONSE_TYPE_IN_CHANNEL, resp.ResponseType)
	assert.Equal(t, model.POST_ME, resp.Type)
	assert.Equal(t, "*"+msg+"*", resp.Text)
}
