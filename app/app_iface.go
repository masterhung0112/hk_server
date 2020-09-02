package app

import (
	"github.com/masterhung0112/go_server/model"
)

type AppIface interface {
  InitServer()
  Session() *model.Session

  // CreateUser creates a user and sets several fields of the returned User struct to
	// their zero values.
  CreateUser(user *model.User) (*model.User, *model.AppError)
  CreateUserWithToken(user *model.User, token *model.Token) (*model.User, *model.AppError)
  CreateUserFromSignup(user *model.User) (*model.User, *model.AppError)
  VerifyUserEmail(userId, email string) *model.AppError

  IsFirstUserAccount() bool
  LimitedClientConfig() map[string]string

}