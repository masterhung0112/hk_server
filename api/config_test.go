package api

import (
  "testing"

	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()
	Client := th.Client

	_, resp := Client.GetConfig()
  CheckForbiddenStatus(t, resp)

  th.TestForSystemAdminAndLocal(t, func(t *testing.T, client *model.Client) {
    cfg, resp := client.GetConfig()
		CheckNoError(t, resp)

		require.NotEqual(t, "", cfg.TeamSettings.SiteName)
  })
}