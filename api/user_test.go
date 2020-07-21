package api

import (
  "testing"

  "github.com/masterhung0112/go_server/model"
)

func TestCreateUser(t *testing.T) {

  user := model.User{
    Id:        "",
    UserName:  "",
    Password:  "",
    Email:     "",
    Roles:     model.SYSTEM_ADMIN_ROLE_ID,
  }


}
