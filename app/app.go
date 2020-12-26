package app

import (
	"context"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/mlog"
	"net/http"
	"strconv"

	"github.com/masterhung0112/hk_server/model"
)

type App struct {
	srv     *Server
	context context.Context

	session model.Session
}

func New(options ...AppOption) *App {
	app := &App{}

	for _, option := range options {
		option(app)
	}

	return app
}

func (a *App) SetContext(c context.Context) {
	a.context = c
}

func (a *App) Srv() *Server {
	return a.srv
}
func (a *App) Log() *mlog.Logger {
	return a.srv.Log
}

func (a *App) Metrics() einterfaces.MetricsInterface {
	return a.srv.Metrics
}

func (a *App) InitServer() {
	a.srv.AppInitializedOnce.Do(func() {
		a.DoAppMigrations()
	})
}

func (a *App) SetServer(srv *Server) {
	a.srv = srv
}

func (a *App) Session() *model.Session {
	return &a.session
}

func (a *App) SetSession(s *model.Session) {
	a.session = *s
}

func (s *Server) getSystemInstallDate() (int64, *model.AppError) {
	systemData, err := s.Store.System().GetByName(model.SYSTEM_INSTALLATION_DATE_KEY)
	if err != nil {
		return 0, model.NewAppError("getSystemInstallDate", "app.system.get_by_name.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	value, err := strconv.ParseInt(systemData.Value, 10, 64)
	if err != nil {
		return 0, model.NewAppError("getSystemInstallDate", "app.system_install_date.parse_int.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return value, nil
}

func (a *App) Cluster() einterfaces.ClusterInterface {
	return a.srv.Cluster
}
