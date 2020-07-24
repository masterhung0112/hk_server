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

  user := model.User{
    UserName:  th.GenerateTestUsername(),
    Password:  "test",
    Email:     th.GenerateTestEmail(),
    Roles:     model.SYSTEM_ADMIN_ROLE_ID,
  }

  ruser, resp := th.Client.CreateUser(&user)
  require.NotNil(t, ruser)
  CheckNoError(t, resp)
  CheckCreatedStatus(t, resp)

  // require.Equal(t, user.Nickname, ruser.Nickname, "nickname didn't match")

}

