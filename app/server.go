package app

import (
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
}

func NewServer(options ...Option) (*Server, error) {
  rootRouter := mux.NewRouter()

  s := &Server{
    RootRouter: rootRouter,
  }

  if s.configStore == nil {
    configStore, err := config.NewFileStore("config.json", true)
    if err != nil {
      return nil, errors.Wrap(err, "failed to load config")
    }

    s.configStore = configStore
  }

  return s, nil
}