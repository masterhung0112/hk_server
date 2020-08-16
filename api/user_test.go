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
}
