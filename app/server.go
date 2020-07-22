package app

import (
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

  return s, nil
}

/*
 * Start HTTP server
 */
func (server *Server) Start() error {

  var handler http.Handler = server.RootRouter

  server.Server = &http.Server{
    Handler:      handler,
  }

  return nil
}