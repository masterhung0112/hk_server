package commands

import (
	"github.com/spf13/cobra"

	"github.com/masterhung0112/hk_server/v5/app"
	"github.com/masterhung0112/hk_server/v5/app/request"
	"github.com/masterhung0112/hk_server/v5/config"
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/shared/i18n"
	"github.com/masterhung0112/hk_server/v5/utils"
)

func initDBCommandContextCobra(command *cobra.Command, readOnlyConfigStore bool) (*app.App, error) {
	a, err := initDBCommandContext(getConfigDSN(command, config.GetEnvironment()), readOnlyConfigStore)
	if err != nil {
		// Returning an error just prints the usage message, so actually panic
		panic(err)
	}

	a.InitPlugins(&request.Context{}, *a.Config().PluginSettings.Directory, *a.Config().PluginSettings.ClientDirectory)
	a.DoAppMigrations()

	return a, nil
}

func InitDBCommandContextCobra(command *cobra.Command) (*app.App, error) {
	return initDBCommandContextCobra(command, true)
}

func InitDBCommandContextCobraReadWrite(command *cobra.Command) (*app.App, error) {
	return initDBCommandContextCobra(command, false)
}

func initDBCommandContext(configDSN string, readOnlyConfigStore bool) (*app.App, error) {
	if err := utils.TranslationsPreInit(); err != nil {
		return nil, err
	}
	model.AppErrorInit(i18n.T)

	s, err := app.NewServer(
		app.Config(configDSN, false, readOnlyConfigStore, nil),
		app.StartSearchEngine,
	)
	if err != nil {
		return nil, err
	}

	a := app.New(app.ServerConnector(s))

	if model.BuildEnterpriseReady == "true" {
		a.Srv().LoadLicense()
	}

	return a, nil
}
