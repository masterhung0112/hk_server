package api1

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/require"
)

func TestGetConfig(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()
	Client := th.Client

	_, resp := Client.GetConfig()
	CheckForbiddenStatus(t, resp)

	th.TestForSystemAdminAndLocal(t, func(t *testing.T, client *model.Client4) {
		cfg, resp := client.GetConfig()
		CheckNoError(t, resp)

		require.NotEqual(t, "", cfg.TeamSettings.SiteName)

		if *cfg.LdapSettings.BindPassword != model.FAKE_SETTING && len(*cfg.LdapSettings.BindPassword) != 0 {
			require.FailNow(t, "did not sanitize properly")
		}
		require.Equal(t, model.FAKE_SETTING, *cfg.FileSettings.PublicLinkSalt, "did not sanitize properly")

		if *cfg.FileSettings.S3SecretAccessKey != model.FAKE_SETTING && len(*cfg.FileSettings.S3SecretAccessKey) != 0 {
			require.FailNow(t, "did not sanitize properly")
		}
		if *cfg.EmailSettings.SMTPPassword != model.FAKE_SETTING && len(*cfg.EmailSettings.SMTPPassword) != 0 {
			require.FailNow(t, "did not sanitize properly")
		}
		// if *cfg.GitLabSettings.Secret != model.FAKE_SETTING && len(*cfg.GitLabSettings.Secret) != 0 {
		// 	require.FailNow(t, "did not sanitize properly")
		// }
		require.Equal(t, model.FAKE_SETTING, *cfg.SqlSettings.DataSource, "did not sanitize properly")
		// require.Equal(t, model.FAKE_SETTING, *cfg.SqlSettings.AtRestEncryptKey, "did not sanitize properly")
		if !strings.Contains(strings.Join(cfg.SqlSettings.DataSourceReplicas, " "), model.FAKE_SETTING) && len(cfg.SqlSettings.DataSourceReplicas) != 0 {
			require.FailNow(t, "did not sanitize properly")
		}
		if !strings.Contains(strings.Join(cfg.SqlSettings.DataSourceSearchReplicas, " "), model.FAKE_SETTING) && len(cfg.SqlSettings.DataSourceSearchReplicas) != 0 {
			require.FailNow(t, "did not sanitize properly")
		}
	})
}

func TestGetConfigWithAccessTag(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	varyByHeader := *&th.App.Config().RateLimitSettings.VaryByHeader // environment perm.
	supportEmail := *&th.App.Config().SupportSettings.SupportEmail   // site perm.
	defer th.App.UpdateConfig(func(cfg *model.Config) {
		cfg.RateLimitSettings.VaryByHeader = varyByHeader
		cfg.SupportSettings.SupportEmail = supportEmail
	})

	// set some values so that we know they're not blank
	mockVaryByHeader := model.NewId()
	mockSupportEmail := model.NewId() + "@hungknow.com"
	th.App.UpdateConfig(func(cfg *model.Config) {
		cfg.RateLimitSettings.VaryByHeader = mockVaryByHeader
		cfg.SupportSettings.SupportEmail = &mockSupportEmail
	})

	th.Client.Login(th.BasicUser.Username, th.BasicUser.Password)

	// add read sysconsole environment config
	th.AddPermissionToRole(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT.Id, model.SYSTEM_USER_ROLE_ID)
	defer th.RemovePermissionFromRole(model.PERMISSION_SYSCONSOLE_READ_ENVIRONMENT.Id, model.SYSTEM_USER_ROLE_ID)

	cfg, resp := th.Client.GetConfig()
	CheckNoError(t, resp)

	t.Run("Cannot read value without permission", func(t *testing.T) {
		assert.Nil(t, cfg.SupportSettings.SupportEmail)
	})

	t.Run("Can read value with permission", func(t *testing.T) {
		assert.Equal(t, mockVaryByHeader, cfg.RateLimitSettings.VaryByHeader)
	})
}
