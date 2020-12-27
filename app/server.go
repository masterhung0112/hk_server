package app

import (
	"context"
	"fmt"
	"github.com/masterhung0112/hk_server/audit"
	"github.com/masterhung0112/hk_server/services/searchengine"
	"github.com/masterhung0112/hk_server/services/searchengine/bleveengine"
	"github.com/masterhung0112/hk_server/store/localcachelayer"
	"github.com/masterhung0112/hk_server/store/retrylayer"
	"hash/maphash"
	"net"
	"net/http"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"

	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/jobs"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/plugin"
	"github.com/masterhung0112/hk_server/services/cache"
	"github.com/masterhung0112/hk_server/services/filesstore"
	"github.com/masterhung0112/hk_server/services/httpservice"
	"github.com/masterhung0112/hk_server/services/imageproxy"
	"github.com/masterhung0112/hk_server/services/timezones"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/store/sqlstore"
	"github.com/masterhung0112/hk_server/utils"
)

var MaxNotificationsPerChannelDefault int64 = 1000000

// declaring this as var to allow overriding in tests
var SENTRY_DSN = "placeholder_sentry_dsn"

type Server struct {
	sqlStore           *sqlstore.SqlStore
	Store              store.Store
	WebSocketRouter    *WebSocketRouter
	AppInitializedOnce sync.Once

	// RootRouter is the starting point for all HTTP requests to the server.
	RootRouter *mux.Router

	// LocalRouter is the starting point for all the local UNIX socket
	// requests to the server
	LocalRouter *mux.Router

	// Router is the starting point for all web, api4 and ws requests to the server. It differs
	// from RootRouter only if the SiteURL contains a /subpath.
	Router *mux.Router

	Server     *http.Server
	ListenAddr *net.TCPAddr
	// RateLimiter *RateLimiter
	Busy *Busy

	localModeServer *http.Server

	didFinishListen chan struct{}

	goroutineCount      int32
	goroutineExitSignal chan struct{}

	PluginsEnvironment     *plugin.Environment
	PluginConfigListenerId string
	PluginsLock            sync.RWMutex

	EmailService *EmailService

	hubs     []*Hub
	hashSeed maphash.Seed

	PushNotificationsHub   PushNotificationsHub
	pushNotificationClient *http.Client // TODO: move this to it's own package

	runjobs bool
	Jobs    *jobs.JobServer

	clusterLeaderListeners sync.Map

	licenseValue       atomic.Value
	clientLicenseValue atomic.Value
	licenseListeners   map[string]func(*model.License, *model.License)

	timezones *timezones.Timezones

	newStore func() (store.Store, error)

	htmlTemplateWatcher     *utils.HTMLTemplateWatcher
	sessionCache            cache.Cache
	seenPendingPostIdsCache cache.Cache
	statusCache             cache.Cache
	configListenerId        string
	licenseListenerId       string
	logListenerId           string
	clusterLeaderListenerId string
	searchConfigListenerId  string
	searchLicenseListenerId string
	loggerLicenseListenerId string
	configStore             *config.Store
	postActionCookieSecret  []byte

	advancedLogListenerCleanup func()

	pluginCommands     []*PluginCommand
	pluginCommandsLock sync.RWMutex

	asymmetricSigningKey atomic.Value
	clientConfig         atomic.Value
	clientConfigHash     atomic.Value
	limitedClientConfig  atomic.Value

	// telemetryService *telemetry.TelemetryService

	phase2PermissionsMigrationComplete bool

	HTTPService httpservice.HTTPService

	ImageProxy *imageproxy.ImageProxy

	Audit            *audit.Audit
	Log              *mlog.Logger
	NotificationsLog *mlog.Logger

	joinCluster       bool
	startMetrics      bool
	startSearchEngine bool

	SearchEngine *searchengine.Broker

	AccountMigration einterfaces.AccountMigrationInterface
	Cluster          einterfaces.ClusterInterface
	Compliance       einterfaces.ComplianceInterface
	DataRetention    einterfaces.DataRetentionInterface
	Ldap             einterfaces.LdapInterface
	MessageExport    einterfaces.MessageExportInterface
	Cloud            einterfaces.CloudInterface
	Metrics          einterfaces.MetricsInterface
	Notification     einterfaces.NotificationInterface
	Saml             einterfaces.SamlInterface

	CacheProvider cache.Provider

	// tracer *tracing.Tracer

	// These are used to prevent concurrent upload requests
	// for a given upload session which could cause inconsistencies
	// and data corruption.
	uploadLockMapMut sync.Mutex
	uploadLockMap    map[string]bool

	// featureFlagSynchronizer      *config.FeatureFlagSynchronizer
	featureFlagStop              chan struct{}
	featureFlagStopped           chan struct{}
	featureFlagSynchronizerMutex sync.Mutex
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
		RootRouter:       rootRouter,
		licenseListeners: map[string]func(*model.License, *model.License){},
		uploadLockMap:    map[string]bool{},
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
		innerStore, err := config.NewFileStore("config.json", true)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load config")
		}
		configStore, err := config.NewStoreFromBacking(innerStore, nil)
		if err != nil {
			return nil, errors.Wrap(err, "failed to load config")
		}

		s.configStore = configStore
	}

	s.HTTPService = httpservice.MakeHTTPService(s)
	s.pushNotificationClient = s.HTTPService.MakeClient(true)

	s.ImageProxy = imageproxy.MakeImageProxy(s, s.HTTPService, s.Log)

	searchEngine := searchengine.NewBroker(s.Config(), s.Jobs)
	bleveEngine := bleveengine.NewBleveEngine(s.Config(), s.Jobs)
	if err := bleveEngine.Start(); err != nil {
		return nil, err
	}
	searchEngine.RegisterBleveEngine(bleveEngine)
	s.SearchEngine = searchEngine

	// at the moment we only have this implementation
	// in the future the cache provider will be built based on the loaded config
	s.CacheProvider = cache.NewProvider()
	if err := s.CacheProvider.Connect(); err != nil {
		return nil, errors.Wrapf(err, "Unable to connect to cache provider")
	}

	var err error
	if s.sessionCache, err = s.CacheProvider.NewCache(&cache.CacheOptions{
		Size:           model.SESSION_CACHE_SIZE,
		Striped:        true,
		StripedBuckets: maxInt(runtime.NumCPU()-1, 1),
	}); err != nil {
		return nil, errors.Wrap(err, "Unable to create session cache")
	}
	if s.seenPendingPostIdsCache, err = s.CacheProvider.NewCache(&cache.CacheOptions{
		Size: PENDING_POST_IDS_CACHE_SIZE,
	}); err != nil {
		return nil, errors.Wrap(err, "Unable to create pending post ids cache")
	}
	if s.statusCache, err = s.CacheProvider.NewCache(&cache.CacheOptions{
		Size:           model.STATUS_CACHE_SIZE,
		Striped:        true,
		StripedBuckets: maxInt(runtime.NumCPU()-1, 1),
	}); err != nil {
		return nil, errors.Wrap(err, "Unable to create status cache")
	}

	s.createPushNotificationsHub()

	// Prepare the translation file
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, errors.Wrapf(err, "unable to load Mattermost translation files")
	}

	if htmlTemplateWatcher, err2 := utils.NewHTMLTemplateWatcher("templates"); err2 != nil {
		mlog.Error("Failed to parse server templates", mlog.Err(err2))
	} else {
		s.htmlTemplateWatcher = htmlTemplateWatcher
	}

	if s.newStore == nil {
		s.newStore = func() (store.Store, error) {
			s.sqlStore = sqlstore.New(s.Config().SqlSettings, s.Metrics)
			if s.sqlStore.DriverName() == model.DATABASE_DRIVER_POSTGRES {
				ver, err2 := s.sqlStore.GetDbVersion(true)
				if err2 != nil {
					return nil, errors.Wrap(err2, "cannot get DB version")
				}
				intVer, err2 := strconv.Atoi(ver)
				if err2 != nil {
					return nil, errors.Wrap(err2, "cannot parse DB version")
				}
				if intVer < sqlstore.MINIMUM_REQUIRED_POSTGRES_VERSION {
					return nil, fmt.Errorf("minimum required postgres version is %s; found %s", sqlstore.VersionString(sqlstore.MINIMUM_REQUIRED_POSTGRES_VERSION), sqlstore.VersionString(intVer))
				}
			}

			lcl, err2 := localcachelayer.NewLocalCacheLayer(
				retrylayer.New(s.sqlStore),
				s.Metrics,
				s.Cluster,
				s.CacheProvider,
			)
			if err2 != nil {
				return nil, errors.Wrap(err2, "cannot create local cache layer")
			}

			searchStore := searchlayer.NewSearchLayer(
				lcl,
				s.SearchEngine,
				s.Config(),
			)

			s.AddConfigListener(func(prevCfg, cfg *model.Config) {
				searchStore.UpdateConfig(cfg)
			})

			s.sqlStore.UpdateLicense(s.License())
			s.AddLicenseListener(func(oldLicense, newLicense *model.License) {
				s.sqlStore.UpdateLicense(newLicense)
			})

			return timerlayer.New(
				searchStore,
				s.Metrics,
			), nil
		}
	}

	s.Store = s.newStore()

	emailService, err := NewEmailService(s)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to initialize email service")
	}
	s.EmailService = emailService

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

	s.timezones = timezones.New()

	license := s.License()
	if license == nil {
		s.UpdateConfig(func(cfg *model.Config) {
			cfg.TeamSettings.MaxNotificationsPerChannel = &MaxNotificationsPerChannelDefault
		})
	}

	s.ReloadConfig()

	allowAdvancedLogging := license != nil && *license.Features.AdvancedLogging

	if s.Audit == nil {
		s.Audit = &audit.Audit{}
		s.Audit.Init(audit.DefMaxQueueSize)
		if err = s.configureAudit(s.Audit, allowAdvancedLogging); err != nil {
			mlog.Error("Error configuring audit", mlog.Err(err))
		}
	}

	mlog.Info("Server is initialized...")

	return s, nil
}

/*
 * Start HTTP server
 */
func (s *Server) Start() error {

	var handler http.Handler = s.RootRouter

	s.Busy = NewBusy(s.Cluster)

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

	if s.htmlTemplateWatcher != nil {
		s.htmlTemplateWatcher.Close()
	}

	if s.advancedLogListenerCleanup != nil {
		s.advancedLogListenerCleanup()
		s.advancedLogListenerCleanup = nil
	}

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
	backend, err := filesstore.NewFileBackend(&s.Config().FileSettings, license != nil && *license.Features.Compliance)
	if err != nil {
		return nil, model.NewAppError("FileBackend", "api.file.no_driver.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return backend, nil
}

func (s *Server) TotalWebsocketConnections() int {
	// This method is only called after the hub is initialized.
	// Therefore, no mutex is needed to protect s.hubs.
	count := int64(0)
	for _, hub := range s.hubs {
		count = count + atomic.LoadInt64(&hub.connectionCount)
	}

	return int(count)
}

func (s *Server) ClientConfigHash() string {
	return s.clientConfigHash.Load().(string)
}

func (s *Server) initJobs() {
	s.Jobs = jobs.NewJobServer(s, s.Store)
	if jobsDataRetentionJobInterface != nil {
		s.Jobs.DataRetentionJob = jobsDataRetentionJobInterface(s)
	}
	if jobsMessageExportJobInterface != nil {
		s.Jobs.MessageExportJob = jobsMessageExportJobInterface(s)
	}
	if jobsElasticsearchAggregatorInterface != nil {
		s.Jobs.ElasticsearchAggregator = jobsElasticsearchAggregatorInterface(s)
	}
	if jobsElasticsearchIndexerInterface != nil {
		s.Jobs.ElasticsearchIndexer = jobsElasticsearchIndexerInterface(s)
	}
	if jobsBleveIndexerInterface != nil {
		s.Jobs.BleveIndexer = jobsBleveIndexerInterface(s)
	}
	if jobsMigrationsInterface != nil {
		s.Jobs.Migrations = jobsMigrationsInterface(s)
	}
}

func (s *Server) TelemetryId() string {
	return ""
	//TODO: Open
	// if s.telemetryService == nil {
	// 	return ""
	// }
	// return s.telemetryService.TelemetryID
}

func (s *Server) HttpService() httpservice.HTTPService {
	return s.HTTPService
}

func (s *Server) SetLog(l *mlog.Logger) {
	s.Log = l
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
