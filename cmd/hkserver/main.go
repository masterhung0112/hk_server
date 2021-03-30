package main

import (
	"os"

	// Enterprise Deps
	_ "github.com/gorilla/handlers"
	_ "github.com/hako/durafmt"
	_ "github.com/hashicorp/memberlist"
	_ "github.com/mattermost/gosaml2"
	_ "github.com/mattermost/ldap"
	_ "github.com/mattermost/rsc/qr"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/tylerb/graceful"
	_ "gopkg.in/olivere/elastic.v6"

	"github.com/masterhung0112/hk_server/v5/cmd/hkserver/commands"
	// Import and register app layer slash commands
	_ "github.com/masterhung0112/hk_server/v5/app/slashcommands"
	// Plugins
	_ "github.com/masterhung0112/hk_server/v5/model/gitlab"
	// Enterprise Imports
	_ "github.com/masterhung0112/hk_server/v5/imports"
)

func main() {
	if err := commands.Run(os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
