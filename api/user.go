package api

import (
  "net/http"
  "github.com/masterhung0112/go_server/model"
  "github.com/masterhung0112/go_server/web"
)

func CreateUser(c *web.Context, w http.ResponseWriter, r *http.Request) {
  // Convert Json to User model
  user := model.UserFromJson(r.Body)

  ruser, err := c.App.CreateUserFromSignup(user)

  if err != nil {
    return
  }

  // Successfully created new user
  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(ruser.ToJson()))
}

func (api *API) InitUser() {
  api.BaseRoutes.Users.Handle("", api.ApiHandler(createUser)).Methods("POST")
  api.BaseRoutes.Users.Handle("", api.ApiHandler(getUsers)).Methods("GET")
}

func createUser(c *Context, w http.ResponseWriter, r *http.Request) {

}

func getUsers(c *Context, w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusOK)
  w.Write([]byte(`{"name": "Hello World" }`))
}