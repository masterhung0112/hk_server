package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/masterhung0112/go_server/config"
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"github.com/masterhung0112/go_server/store/sqlstore"
	"github.com/masterhung0112/go_server/utils"
	"github.com/pkg/errors"
)

type Server struct {
	newStore    func() store.Store
	sqlStore    *sqlstore.SqlSupplier
	Store       store.Store
	configStore config.Store

	didFinishListen chan struct{}

	// RootRouter is the starting point for all HTTP requests to the server.
	RootRouter *mux.Router

	// LocalRouter is the starting point for all the local UNIX socket
	// requests to the server
	LocalRouter *mux.Router

	// Router is the starting point for all web, api4 and ws requests to the server.
	// It differs from RootRouter only if the SiteURL contains a /subpath
	Router *mux.Router

	Server     *http.Server
	ListenAddr *net.TCPAddr

	clientConfig        atomic.Value
	clientConfigHash    atomic.Value
	limitedClientConfig atomic.Value

	postActionCookieSecret []byte
}

// Global app options that should be applied to apps created by this server
func (s *Server) AppOptions() []AppOption {
	return []AppOption{
		ServerConnector(s),
	}
}

func NewServer(options ...Option) (*Server, error) {
	rootRouter := mux.NewRouter()

	mlog.Info("Server is initializing...")
	// Create Server instance
	s := &Server{
		RootRouter: rootRouter,
	}

	for _, option := range options {
		if err := option(s); err != nil {
			return nil, errors.Wrap(err, "failed to apply option")
		}
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

	if s.newStore == nil {
		s.newStore = func() store.Store {
			s.sqlStore = sqlstore.NewSqlSupplier(s.Config().SqlSettings)
			return s.sqlStore
		}
	}

	s.Store = s.newStore()

	if err := s.ensureInstallationDate(); err != nil {
		return nil, errors.Wrapf(err, "unable to ensure installation date")
	}

	s.regenerateClientConfig()

	// Prepare Router for all Web, WS paths
	subpath, err := utils.GetSubpathFromConfig(s.Config())
	if err != nil {
		return nil, errors.Wrap(err, "Failed to parse SiteURL subpath")
	}
	s.Router = s.RootRouter.PathPrefix(subpath).Subrouter()

	return s, nil
}

/*
 * Start HTTP server
 */
func (s *Server) Start() error {

	var handler http.Handler = s.RootRouter

	// Set HTTP handler of gozilla
	s.Server = &http.Server{
		Handler: handler,
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

	logListeningPort := fmt.Sprintf("Server is listening on %v", listener.Addr().String())
	mlog.Info(logListeningPort, mlog.String("address", listener.Addr().String()))

	s.didFinishListen = make(chan struct{})
	go func() {
		err = s.Server.Serve(listener)

		if err != nil && err != http.ErrServerClosed {
			mlog.Critical("Error starting server", mlog.Err(err))
			time.Sleep(time.Second)
		}

		close(s.didFinishListen)
	}()

	return nil
}

func (s *Server) Shutdown() error {
	mlog.Info("Stopping Server...")

	s.StopHttpServer()

	if s.Store != nil {
		s.Store.Close()
	}

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
