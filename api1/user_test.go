package api1

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/masterhung0112/hk_server/model"
)

func TestConnectUserGet(t *testing.T) {
	th := Setup(t)

	_, err := th.Client.DoApiGet("/users", "")
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestCreateUser(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	user := model.User{
		Username: GenerateTestUsername(),
		Password: "hello1",
		Email:    th.GenerateTestEmail(),
		Roles:    model.SYSTEM_ADMIN_ROLE_ID,
	}

	ruser, resp := th.Client.CreateUser(&user)
	CheckNoError(t, resp)
	CheckCreatedStatus(t, resp)
	require.NotNil(t, ruser)

	_, _ = th.Client.Login(user.Email, user.Password)

	// require.Equal(t, user.Nickname, ruser.Nickname, "nickname didn't match")
	require.Equal(t, model.SYSTEM_USER_ROLE_ID, ruser.Roles, "did not clear roles")

	CheckUserSanitization(t, ruser)

	_, resp = th.Client.CreateUser(ruser)
	CheckBadRequestStatus(t, resp)

	ruser.Id = ""
	ruser.Username = GenerateTestUsername()
	ruser.Password = "passwd1"
	_, resp = th.Client.CreateUser(ruser)
	CheckErrorMessage(t, resp, "app.user.save.email_exists.app_error")
	CheckBadRequestStatus(t, resp)

	ruser.Email = th.GenerateTestEmail()
	ruser.Username = user.Username
	_, resp = th.Client.CreateUser(ruser)
	CheckErrorMessage(t, resp, "app.user.save.username_exists.app_error")
	CheckBadRequestStatus(t, resp)

	ruser.Email = ""
	_, resp = th.Client.CreateUser(ruser)
	CheckErrorMessage(t, resp, "model.user.is_valid.email.app_error")
	CheckBadRequestStatus(t, resp)

	ruser.Username = "testinvalid+++"
	_, resp = th.Client.CreateUser(ruser)
	CheckErrorMessage(t, resp, "model.user.is_valid.username.app_error")
	CheckBadRequestStatus(t, resp)

	// th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.EnableOpenServer = false })
	// th.App.UpdateConfig(func(cfg *model.Config) { *cfg.TeamSettings.EnableUserCreation = false })

	// th.TestForSystemAdminAndLocal(t, func(t *testing.T, client *model.Client1) {
	// 	user2 := &model.User{Email: th.GenerateTestEmail(), Password: "Password1", Username: GenerateTestUsername()}
	// 	_, resp = client.CreateUser(user2)
	// 	CheckNoError(t, resp)

	// 	r, err := client.DoApiPost("/users", "garbage")
	// 	require.NotNil(t, err, "should have errored")
	// 	assert.Equal(t, http.StatusBadRequest, r.StatusCode)
	// })
}

func TestGetMe(t *testing.T) {
	th := Setup(t).InitBasic()
	defer th.TearDown()

	ruser, resp := th.Client.GetMe("")
	CheckNoError(t, resp)

	require.Equal(t, th.BasicUser.Id, ruser.Id)

	th.Client.Logout()
	_, resp = th.Client.GetMe("")
	CheckUnauthorizedStatus(t, resp)
}

func TestGetUsers(t *testing.T) {
	th := Setup(t)
	defer th.TearDown()

	th.TestForAllClients(t, func(t *testing.T, client *model.Client1) {
		rusers, resp := client.GetUsers(0, 60, "")
		CheckNoError(t, resp)
		for _, u := range rusers {
			CheckUserSanitization(t, u)
		}

		rusers, resp = client.GetUsers(0, 1, "")
		CheckNoError(t, resp)
		require.Len(t, rusers, 1, "should be 1 per page")

		rusers, resp = client.GetUsers(1, 1, "")
		CheckNoError(t, resp)
		require.Len(t, rusers, 1, "should be 1 per page")

		rusers, resp = client.GetUsers(10000, 100, "")
		CheckNoError(t, resp)
		require.Empty(t, rusers, "should be no users")

		// Check default params for page and per_page
		_, err := client.DoApiGet("/users", "")
		require.Nil(t, err)
	})

	th.Client.Logout()
	_, resp := th.Client.GetUsers(0, 60, "")
	CheckUnauthorizedStatus(t, resp)
}
