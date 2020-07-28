package commands

import (
	"syscall"
	"os"
	"os/signal"
	"github.com/masterhung0112/go_server/api"
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/app"
	"github.com/masterhung0112/go_server/config"
	"github.com/spf13/cobra"

)

var serverCmd = &cobra.Command{
  Use: "server",
  Short: "Run the server",
  RunE: serverCmdF,
  SilenceUsage: true,
}

func init() {
  RootCmd.AddCommand(serverCmd)
  RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
  return nil
}

func runServer(configStore config.Store) error {
  options := []app.Option{
    app.ConfigStore(configStore),
  }
  server, err := app.NewServer(options...)
  if err != nil {
		mlog.Critical(err.Error())
		return err
  }
  defer server.Shutdown()

  // server, server.AppOptions,
  api.ApiInit(server.Router)

  serverErr := server.Start()
  if serverErr != nil {
    mlog.Critical(serverErr.Error())
    return serverErr
  }

  interruptChan := make(chan os.Signal, 1)

  // wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

  return nil
}