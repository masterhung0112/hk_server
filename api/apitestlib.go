package api

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/masterhung0112/go_server/app"
	"github.com/masterhung0112/go_server/config"
	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	App         *app.App
	Server      *app.Server
	ConfigStore config.Store
	Client      *model.Client

	SystemAdminClient *model.Client
	LocalClient       *model.Client
}

func setupTestHelper() *TestHelper {
	var options []app.Option

	memoryStore, err := config.NewMemoryStoreWithOptions(&config.MemoryStoreOptions{IgnoreEnvironmentOverrides: true})
	if err != nil {
		panic("failed to initialize memory store: " + err.Error())
	}

	config := memoryStore.Get()
	// *config.PluginSettings.Directory = filepath.Join(tempWorkspace, "plugins")
	// *config.PluginSettings.ClientDirectory = filepath.Join(tempWorkspace, "webapp")
	// config.ServiceSettings.EnableLocalMode = model.NewBool(true)
	// *config.ServiceSettings.LocalModeSocketLocation = filepath.Join(tempWorkspace, "mattermost_local.sock")
	// if updateConfig != nil {
	// 	updateConfig(config)
	// }
	memoryStore.Set(config)

	options = append(options, app.ConfigStore(memoryStore))

	s, err := app.NewServer(options...)
	if err != nil {
		panic(err)
	}

	th := &TestHelper{
		App:    app.New(app.ServerConnector(s)),
		Server: s,
	}

	// Initialize the router URL
	ApiInit(th.Server.AppOptions, th.App.Srv().Router)

	// Start HTTP Server and other stuff
	if err := th.Server.Start(); err != nil {
		panic(err)
	}

	// Disable strict password requirements for test
	th.App.UpdateConfig(func(cfg *model.Config) {
		*cfg.PasswordSettings.MinimumLength = 5
		*cfg.PasswordSettings.Lowercase = false
		*cfg.PasswordSettings.Uppercase = false
		*cfg.PasswordSettings.Symbol = false
		*cfg.PasswordSettings.Number = false
	})

	th.Client = th.CreateClient()
  th.SystemAdminClient = th.CreateClient()
  //TODO: Open this
	// th.LocalClient = th.CreateLocalClient(*config.ServiceSettings.LocalModeSocketLocation)

	th.App.InitServer()

	return th
}

func (th *TestHelper) CreateClient() *model.Client {
	return model.NewApiClient(fmt.Sprintf("http://localhost:%v", th.App.Srv().ListenAddr.Port))
}

func Setup(tb testing.TB) *TestHelper {
	return setupTestHelper()
}

func (me *TestHelper) TearDown() {
	// utils.DisableDebugLogForTest()
	// if me.IncludeCacheLayer {
	// 	// Clean all the caches
	// 	me.App.Srv().InvalidateAllCaches()
	// }

	me.ShutdownApp()

	// utils.EnableDebugLogForTest()
}

func (me *TestHelper) ShutdownApp() {
	done := make(chan bool)
	go func() {
		me.Server.Shutdown()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(30 * time.Second):
		// panic instead of fatal to terminate all tests in this package, otherwise the
		// still running App could spuriously fail subsequent tests.
		panic("failed to shutdown App within 30 seconds")
	}
}

// ToDo: maybe move this to NewAPIv4SocketClient and reuse it in mmctl
func (me *TestHelper) CreateLocalClient(socketPath string) *model.Client {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
	}

	return &model.Client{
		ApiUrl:     "http://_" + model.API_URL_SUFFIX,
		HttpClient: httpClient,
	}
}

func (th *TestHelper) GenerateTestEmail() string {
	return strings.ToLower(model.NewId() + "@localhost")
}

func (th *TestHelper) GenerateTestUsername() string {
	return "fakeuser" + model.NewRandomString(10)
}

func checkHTTPStatus(t *testing.T, resp *model.Response, expectedStatus int, expectError bool) {
	t.Helper()

	require.NotNilf(t, resp, "Unexpected nil response, expected http:%v, expectError:%v", expectedStatus, expectError)
	if expectError {
		require.NotNil(t, resp.Error, "Expected a non-nil error and http status:%v, got nil, %v, error: %v", expectedStatus, resp.StatusCode)
	} else {
		require.Nil(t, resp.Error, "Expected no error and http status:%v, got %q, http:%v", expectedStatus, resp.Error, resp.StatusCode)
	}
	require.Equalf(t, expectedStatus, resp.StatusCode, "Expected http status:%v, got %v (err: %q)", expectedStatus, resp.StatusCode, resp.Error)
}

func CheckNoError(t *testing.T, resp *model.Response) {
	t.Helper()

	require.Nil(t, resp.Error, "expected no error")
}

func CheckCreatedStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusCreated, false)
}

func CheckBadRequestStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusBadRequest, true)
}

func CheckUserSanitization(t *testing.T, user *model.User) {
	t.Helper()

	require.Equal(t, "", user.Password, "password wasn't blank")
	//TODO: Open
	// require.Empty(t, user.AuthData, "auth data wasn't blank")
	// require.Equal(t, "", user.MfaSecret, "mfa secret wasn't blank")
}

func CheckErrorMessage(t *testing.T, resp *model.Response, errorId string) {
	t.Helper()

	require.NotNilf(t, resp.Error, "should have errored with message: %s", errorId)
	require.Equalf(t, resp.Error.Id, errorId, "incorrect error message, actual: %s, expected: %s", resp.Error.Id, errorId)
}

func GenerateTestUsername() string {
	return "fakeuser" + model.NewRandomString(10)
}

// TestForAllClients runs a test function for all the clients
// registered in the TestHelper
func (me *TestHelper) TestForAllClients(t *testing.T, f func(*testing.T, *model.Client), name ...string) {
	var testName string
	if len(name) > 0 {
		testName = name[0] + "/"
	}

	t.Run(testName+"Client", func(t *testing.T) {
		f(t, me.Client)
	})

	t.Run(testName+"SystemAdminClient", func(t *testing.T) {
		f(t, me.SystemAdminClient)
	})

  //TODO: Open this
	// t.Run(testName+"LocalClient", func(t *testing.T) {
	// 	f(t, me.LocalClient)
	// })
}

func CheckUnauthorizedStatus(t *testing.T, resp *model.Response) {
	t.Helper()
	checkHTTPStatus(t, resp, http.StatusUnauthorized, true)
}
