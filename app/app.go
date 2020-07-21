package app

import "context"

type App struct {
  srv     *Server
  context context.Context
}

func (a *App) SetContext(c context.Context) {
  a.context = c
}

func (a *App) Srv() *Server {
  return a.srv
}