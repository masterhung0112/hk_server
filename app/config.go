package app

import(
  "github.com/masterhung0112/go_server/model"
)

func (s *Server) Config() *model.Config {
  return s.configStore.Get()
}

func (a *App) Config() *model.Config {
  return a.Srv().Config()
}