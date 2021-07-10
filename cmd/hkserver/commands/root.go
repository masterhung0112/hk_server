package commands

import (
	"github.com/spf13/cobra"
)

type Command = cobra.Command

func Run(args []string) error {
	RootCmd.SetArgs(args)
	return RootCmd.Execute()
}

var RootCmd = &cobra.Command{
	Use:   "hkserver",
	Short: "Self-hosted server",
	Long:  `Mattermost offers workplace messaging across web, PC and phones with archiving, search and integration with your existing systems. Documentation available at https://docs.mattermost.com`,
}

func init() {
	RootCmd.PersistentFlags().StringP("config", "c", "", "Configuration file to use.")
	RootCmd.PersistentFlags().Bool("disableconfigwatch", false, "When set config.json will not be loaded from disk when the file is changed.")
}
