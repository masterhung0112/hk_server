package api

import (
  "net/http"

  "github.com/masterhung0112/go_server/web"
)

func CreateUser(c *web.Context, w http.ResponseWriter, r *http.Request) {
  user := model.UserFromJson(r.Body)

  ruser, err := c.App.CreateUserFromSignup(user)
}