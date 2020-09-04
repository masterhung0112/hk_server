package commands

import (
	"github.com/pkg/errors"
	"github.com/masterhung0112/go_server/utils"
	"github.com/mattermost/viper"
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
  SilenceUsage: false,
}

func init() {
  RootCmd.AddCommand(serverCmd)
  RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
  configDSN := viper.GetString("config")

  disableConfigWatch, _ := command.Flags().GetBool("disableconfigwatch")
  //TODO: open this
	// usedPlatform, _ := command.Flags().GetBool("platform")

  interruptChan := make(chan os.Signal, 1)
  if err := utils.TranslationsPreInit(); err != nil {
		return errors.Wrapf(err, "unable to load Mattermost translation files")
	}
	configStore, err := config.NewStore(configDSN, !disableConfigWatch)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration")
	}

  //TODO: Replace this
	return runServer(configStore, interruptChan) //runServer(configStore, disableConfigWatch, usedPlatform, interruptChan)
}

func runServer(configStore config.Store, interruptChan chan os.Signal) error {
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
  api.ApiInit(server.AppOptions, server.Router)

  serverErr := server.Start()
  if serverErr != nil {
    mlog.Critical(serverErr.Error())
    return serverErr
  }

  // wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

  return nil
}