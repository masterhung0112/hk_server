package api

import (
	"github.com/stretchr/testify/require"
  "testing"

  "github.com/masterhung0112/go_server/model"
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
    Username:  th.GenerateTestUsername(),
    Password:  "hello1",
    Email:     th.GenerateTestEmail(),
    Roles:     model.SYSTEM_ADMIN_ROLE_ID,
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
	CheckErrorMessage(t, resp, "store.sql_user.save.email_exists.app_error")
  CheckBadRequestStatus(t, resp)

  ruser.Email = th.GenerateTestEmail()
	ruser.Username = user.Username
	_, resp = th.Client.CreateUser(ruser)
	CheckErrorMessage(t, resp, "store.sql_user.save.username_exists.app_error")
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

	// th.TestForSystemAdminAndLocal(t, func(t *testing.T, client *model.Client4) {
	// 	user2 := &model.User{Email: th.GenerateTestEmail(), Password: "Password1", Username: GenerateTestUsername()}
	// 	_, resp = client.CreateUser(user2)
	// 	CheckNoError(t, resp)

	// 	r, err := client.DoApiPost("/users", "garbage")
	// 	require.NotNil(t, err, "should have errored")
	// 	assert.Equal(t, http.StatusBadRequest, r.StatusCode)
	// })
}
