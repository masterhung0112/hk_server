package app

import (
	"context"

	"github.com/masterhung0112/go_server/model"
)

type App struct {
	srv           *Server
  context       context.Context

  session       model.Session
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

func (a *App) InitServer() {
	//TODO: Add implementation
}

func (a *App) SetServer(srv *Server) {
	a.srv = srv
}

func (a *App) Session() *model.Session {
	return &a.session
}
