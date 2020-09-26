package app

import (
	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

type Option func(server *Server) error

type AppOption func(a *App)
type AppOptionCreator func() []AppOption

func ServerConnector(s *Server) AppOption {
	return func(a *App) {
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

// By default, the app will use the store specified by the configuration. This allows you to
// construct an app with a different store.
//
// The override parameter must be either a store.Store or func(App) store.Store().
func StoreOverride(override interface{}) Option {
	return func(s *Server) error {
		switch o := override.(type) {
		case store.Store:
			s.newStore = func() store.Store {
				return o
			}
			return nil

		case func(*Server) store.Store:
			s.newStore = func() store.Store {
				return o(s)
			}
			return nil

		default:
			return errors.New("invalid StoreOverride")
		}
	}
}
