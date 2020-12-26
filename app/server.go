package app

import (
	"context"
	"fmt"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/plugin"
	"github.com/masterhung0112/hk_server/services/filesstore"
	"github.com/masterhung0112/hk_server/services/httpservice"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/store/sqlstore"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/pkg/errors"
)

type Server struct {
	newStore    func() store.Store
	sqlStore    *sqlstore.SqlStore
	Store       store.Store
	configStore config.Store

	licenseValue atomic.Value

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

	goroutineCount      int32
	goroutineExitSignal chan struct{}

	AppInitializedOnce sync.Once

	phase2PermissionsMigrationComplete bool

	Cluster einterfaces.ClusterInterface
	Metrics einterfaces.MetricsInterface

	HTTPService httpservice.HTTPService

	Log *mlog.Logger

	PluginsEnvironment     *plugin.Environment
	PluginConfigListenerId string
	PluginsLock            sync.RWMutex
}

// Global app options that should be applied to apps created by this server
func (s *Server) AppOptions() []AppOption {
	return []AppOption{
		ServerConnector(s),
	}
}

// initLogging initializes and configures the logger. This may be called more than once.
func (s *Server) initLogging() error {
	if s.Log == nil {
		s.Log = mlog.NewLogger(utils.MloggerConfigFromLoggerConfig(&s.Config().LogSettings, utils.GetLogFileLocation))
	}

	// Use this app logger as the global logger (eventually remove all instances of global logging).
	// This is deferred because a copy is made of the logger and it must be fully configured before
	// the copy is made.
	defer mlog.InitGlobalLogger(s.Log)

	// Redirect default Go logger to this logger.
	defer mlog.RedirectStdLog(s.Log)

	if s.NotificationsLog == nil {
		notificationLogSettings := utils.GetLogSettingsFromNotificationsLogSettings(&s.Config().NotificationLogSettings)
		s.NotificationsLog = mlog.NewLogger(utils.MloggerConfigFromLoggerConfig(notificationLogSettings, utils.GetNotificationsLogFileLocation)).
			WithCallerSkip(1).With(mlog.String("logSource", "notifications"))
	}

	if s.logListenerId != "" {
		s.RemoveConfigListener(s.logListenerId)
	}
	s.logListenerId = s.AddConfigListener(func(_, after *model.Config) {
		s.Log.ChangeLevels(utils.MloggerConfigFromLoggerConfig(&after.LogSettings, utils.GetLogFileLocation))

		notificationLogSettings := utils.GetLogSettingsFromNotificationsLogSettings(&after.NotificationLogSettings)
		s.NotificationsLog.ChangeLevels(utils.MloggerConfigFromLoggerConfig(notificationLogSettings, utils.GetNotificationsLogFileLocation))
	})

	// Configure advanced logging.
	// Advanced logging is E20 only, however logging must be initialized before the license
	// file is loaded.  If no valid E20 license exists then advanced logging will be
	// shutdown once license is loaded/checked.
	if *s.Config().LogSettings.AdvancedLoggingConfig != "" {
		dsn := *s.Config().LogSettings.AdvancedLoggingConfig
		isJson := config.IsJsonMap(dsn)

		// If this is a file based config we need the full path so it can be watched.
		if !isJson && strings.HasPrefix(s.configStore.String(), "file://") && !filepath.IsAbs(dsn) {
			configPath := strings.TrimPrefix(s.configStore.String(), "file://")
			dsn = filepath.Join(filepath.Dir(configPath), dsn)
		}

		cfg, err := config.NewLogConfigSrc(dsn, isJson, s.configStore)
		if err != nil {
			return fmt.Errorf("invalid advanced logging config, %w", err)
		}

		if err := s.Log.ConfigAdvancedLogging(cfg.Get()); err != nil {
			return fmt.Errorf("error configuring advanced logging, %w", err)
		}

		if !isJson {
			mlog.Info("Loaded advanced logging config", mlog.String("source", dsn))
		}

		listenerId := cfg.AddListener(func(_, newCfg mlog.LogTargetCfg) {
			if err := s.Log.ConfigAdvancedLogging(newCfg); err != nil {
				mlog.Error("Error re-configuring advanced logging", mlog.Err(err))
			} else {
				mlog.Info("Re-configured advanced logging")
			}
		})

		// In case initLogging is called more than once.
		if s.advancedLogListenerCleanup != nil {
			s.advancedLogListenerCleanup()
		}

		s.advancedLogListenerCleanup = func() {
			cfg.RemoveListener(listenerId)
		}
	}
	return nil
}

func NewServer(options ...Option) (*Server, error) {
	rootRouter := mux.NewRouter()

	// Create Server instance
	s := &Server{
		RootRouter: rootRouter,
	}

	if err := s.initLogging(); err != nil {
		mlog.Error(err.Error())
	}

	mlog.Info("Server is initializing...")

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

	s.HTTPService = httpservice.MakeHTTPService(s)
	// s.pushNotificationClient = s.HTTPService.MakeClient(true)

	// s.ImageProxy = imageproxy.MakeImageProxy(s, s.HTTPService, s.Log)

	// Prepare the translation file
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, errors.Wrapf(err, "unable to load Mattermost translation files")
	}

	if s.newStore == nil {
		s.newStore = func() store.Store {
			s.sqlStore = sqlstore.New(s.Config().SqlSettings, s.Metrics)
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

	mlog.Info("Server is initialized...")

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

// Go creates a goroutine, but maintains a record of it to ensure that execution completes before
// the server is shutdown.
func (s *Server) Go(f func()) {
	atomic.AddInt32(&s.goroutineCount, 1)

	go func() {
		f()

		atomic.AddInt32(&s.goroutineCount, -1)
		select {
		case s.goroutineExitSignal <- struct{}{}:
		default:
		}
	}()
}

func (s *Server) FileBackend() (filesstore.FileBackend, *model.AppError) {
	license := s.License()
	return filesstore.NewFileBackend(&s.Config().FileSettings, license != nil && *license.Features.Compliance)
}

func (s *Server) HttpService() httpservice.HTTPService {
	return s.HTTPService
}

func (s *Server) SetLog(l *mlog.Logger) {
	s.Log = l
}
