package api

import (
	"github.com/gorilla/mux"
	"github.com/masterhung0112/go_server/app"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/web"
)

type ApiRoutes struct {
	Root    *mux.Router // ''
	ApiRoot *mux.Router // 'api/v1'

	Users *mux.Router // 'api/v1/users'
	User  *mux.Router // 'api/v1/users/{user_id:[A-Za-z0-9]+}'

	Teams *mux.Router // 'api/v1/teams'
}

type API struct {
	BaseRoutes          *ApiRoutes
	GetGlobalAppOptions app.AppOptionCreator
}

func ApiInit(globalOptionsFunc app.AppOptionCreator, root *mux.Router) *API {
	api := &API{
		BaseRoutes:          &ApiRoutes{},
		GetGlobalAppOptions: globalOptionsFunc,
	}

	api.BaseRoutes.Root = root
	api.BaseRoutes.ApiRoot = root.PathPrefix(model.API_URL_SUFFIX).Subrouter()

	api.BaseRoutes.Users = api.BaseRoutes.ApiRoot.PathPrefix("/users").Subrouter()
	api.BaseRoutes.User = api.BaseRoutes.ApiRoot.PathPrefix("/users/{user_id:[A-za-z0-9]+}").Subrouter()

	api.BaseRoutes.Teams = api.BaseRoutes.ApiRoot.PathPrefix("/teams").Subrouter()
	api.InitUser()
	api.InitTeam()
	api.InitConfig()

	return api
}

var ReturnStatusOK = web.ReturnStatusOK
