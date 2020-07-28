package app

import (
	"github.com/masterhung0112/go_server/config"
)

type Option func(server *Server) error

type AppOption func(a *App)
type AppOptionCreator func() []AppOption

func ServerConnector(s *Server) AppOption {
  return func (a *App) {
    a.srv = s
  }
}

// ConfigStore applies the given config store,
// typically to replace the traditional sources with a memory store for testing
func ConfigStore(configStore config.Store) Option {
  return func(s *Server) error {
    s.configStore = configStore
    return nil
  }
}