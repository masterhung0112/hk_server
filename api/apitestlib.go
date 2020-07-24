package api

import (
	"net/http"
	"github.com/stretchr/testify/require"
	"strings"
	"fmt"
	"testing"
	"github.com/masterhung0112/go_server/config"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/app"

)

type TestHelper struct {
  App         *app.App
  Server      *app.Server
  ConfigStore config.Store
  Client      *model.Client
}

func setupTestHelper() *TestHelper {
  var options []app.Option

  s, err := app.NewServer(options...)
  if err != nil {
    panic(err)
  }

  th := &TestHelper {
    App:    app.New(app.ServerConnector(s)),
    Server: s,
  }

  // Initialize the router URL
  ApiInit(th.App.Srv().Router)

  // Start HTTP Server and other stuff
  if err := th.Server.Start(); err != nil {
    panic(err)
  }

  th.Client = th.CreateClient()

  return th
}

func (th *TestHelper) CreateClient() *model.Client {
  return model.NewApiClient(fmt.Sprintf("http://localhost:%v", th.App.Srv().ListenAddr.Port))
}

func Setup(tb testing.TB) *TestHelper {
  return setupTestHelper()
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
		require.NotNil(t, resp.Error, "Expected a non-nil error and http status:%v, got nil, %v", expectedStatus, resp.StatusCode)
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