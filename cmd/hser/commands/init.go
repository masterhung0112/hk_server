package commands

import (
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/mattermost/viper"
	"github.com/spf13/cobra"
)

func InitDBCommandContextCobra(command *cobra.Command) (*app.App, error) {
	config := viper.GetString("config")

	a, err := InitDBCommandContext(config)

	if err != nil {
		// Returning an error just prints the usage message, so actually panic
		panic(err)
	}

	//TODO: Open
	// a.InitPlugins(*a.Config().PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory)
	a.DoAppMigrations()

	return a, nil
}

func InitDBCommandContext(configDSN string) (*app.App, error) {
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, err
	}
	model.AppErrorInit(utils.T)

	s, err := app.NewServer(
		app.Config(configDSN, false),
		// app.StartSearchEngine,
	)
	if err != nil {
		return nil, err
	}

	a := app.New(app.ServerConnector(s))

	if model.BuildEnterpriseReady == "true" {
		//TODO: Open
		// a.Srv().LoadLicense()
	}

	return a, nil
}
