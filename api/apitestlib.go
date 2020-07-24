package api

import (
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