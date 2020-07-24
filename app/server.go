package app

import (
	"github.com/masterhung0112/go_server/utils"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/mlog"
	"net/http"
	"net"
  "github.com/pkg/errors"
  "github.com/gorilla/mux"
  "github.com/masterhung0112/go_server/config"
)

type Server struct {
  configStore             config.Store

  // RootRouter is the starting point for all HTTP requests to the server.
  RootRouter *mux.Router

  // LocalRouter is the starting point for all the local UNIX socket
  // requests to the server
  LocalRouter *mux.Router

  Server      *http.Server
  ListenAddr  *net.TCPAddr
}

func NewServer(options ...Option) (*Server, error) {
  rootRouter := mux.NewRouter()

  mlog.Info("Server is initializing...")
  // Create Server instance
  s := &Server{
    RootRouter: rootRouter,
  }

  // Read config from config.json file
  if s.configStore == nil {
    configStore, err := config.NewFileStore("config.json", true)
    if err != nil {
      return nil, errors.Wrap(err, "failed to load config")
    }

    s.configStore = configStore
  }

  if err := utils.TranslationsPreInit(); err != nil {
    return nil, errors.Wrapf(err, "unable to load Mattermost translation files")
  }

  return s, nil
}

/*
 * Start HTTP server
 */
func (s *Server) Start() error {

  var handler http.Handler = s.RootRouter

  s.Server = &http.Server{
    Handler:      handler,
  }

  addr := *s.Config().ServiceSettings.ListenAddress
	if addr == "" {
		if *s.Config().ServiceSettings.ConnectionSecurity == model.CONN_SECURITY_TLS {
			addr = ":https"
		} else {
			addr = ":http"
		}
  }

  listener, err := net.Listen("tcp", addr)
  if err != nil {
    errors.Wrap(err, utils.T("api.server.start_server.starting.critical", err))
    return err
  }
  s.ListenAddr = listener.Addr().(*net.TCPAddr)

  go func() {
    err = s.Server.Serve(listener)
  }()

  return nil
}

func (server *Server) Shutdown() error {
  return nil
}

func (s *Server) UpdateConfig(f func(*model.Config)) {
  old := s.Config()
  updated := old.Clone()
  f(updated)
  if _, err := s.configStore.Set(updated); err != nil {
    mlog.Error("Failed to update config", mlog.Err(err))
  }
}