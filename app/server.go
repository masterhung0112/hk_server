package app

import (
	"time"
	"context"
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

  didFinishListen chan struct{}

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

  // Prepare the translation file
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

  // Get IP and port for listening
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

  s.didFinishListen = make(chan struct{})
  go func() {
    err = s.Server.Serve(listener)

    close(s.didFinishListen)
  }()

  return nil
}

func (s *Server) Shutdown() error {
  mlog.Info("Stopping Server...")

  s.StopHttpServer()

  // if s.Store != nil {
  //   s.Store.Close()
  // }

  mlog.Info("Stopped Server...")

  return nil
}

const TIME_TO_WAIT_FOR_CONNECTIONS_TO_CLOSE_ON_SERVER_SHUTDOWN = time.Second

func (s *Server) StopHttpServer() {
  if s.Server != nil {
    ctx, cancel := context.WithTimeout(context.Background(), TIME_TO_WAIT_FOR_CONNECTIONS_TO_CLOSE_ON_SERVER_SHUTDOWN)
    defer cancel()

    didShutdown := false
    for s.didFinishListen != nil && !didShutdown {
      if err := s.Server.Shutdown(ctx); err != nil {
        mlog.Warn("Unable to shutdown server", mlog.Err(err))
      }
      timer := time.NewTimer(time.Millisecond * 50)
      select {
      case <-s.didFinishListen:
        didShutdown = true
      case <-timer.C:
      }
      timer.Stop()
    }
    s.Server.Close()
    s.Server = nil
  }
}

func (s *Server) UpdateConfig(f func(*model.Config)) {
  old := s.Config()
  updated := old.Clone()
  f(updated)
  if _, err := s.configStore.Set(updated); err != nil {
    mlog.Error("Failed to update config", mlog.Err(err))
  }
}