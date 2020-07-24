package api

import (
	"github.com/stretchr/testify/require"
  "testing"

  // "github.com/masterhung0112/go_server/model"
)

func TestConnectUserGet(t *testing.T) {
  th := Setup(t)

  r, err := th.Client.DoApiGet("/articles/users/", "")
  if err != nil {
    t.Fatal(err.Error())
  }
}


func TestCreateUser(t *testing.T) {

  // user := model.User{
  //   Id:        "",
  //   UserName:  "",
  //   Password:  "",
  //   Email:     "",
  //   Roles:     model.SYSTEM_ADMIN_ROLE_ID,
  // }


}
