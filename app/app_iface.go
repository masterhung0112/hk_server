package app

import (
	"github.com/masterhung0112/hk_server/model"
	"net/http"
)

type AppIface interface {
	InitServer()
	Session() *model.Session
	Srv() *Server

	// CreateUser creates a user and sets several fields of the returned User struct to
	// their zero values.
	CreateUser(user *model.User) (*model.User, *model.AppError)
	CreateUserWithToken(user *model.User, token *model.Token) (*model.User, *model.AppError)
	CreateUserFromSignup(user *model.User, redirect string) (*model.User, *model.AppError)
	VerifyUserEmail(userId, email string) *model.AppError

	IsFirstUserAccount() bool
	LimitedClientConfig() map[string]string

	GetSanitizeOptions(asAdmin bool) map[string]bool
	Config() *model.Config
	Handle404(w http.ResponseWriter, r *http.Request)
	NotifySessionsExpired() *model.AppError
	UpdateProductNotices() *model.AppError
}
