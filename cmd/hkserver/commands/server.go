package commands

import (
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	api4 "github.com/masterhung0112/hk_server/v5/api4"
	"github.com/masterhung0112/hk_server/v5/app"
	"github.com/masterhung0112/hk_server/v5/config"
	"github.com/masterhung0112/hk_server/v5/manualtesting"
	"github.com/masterhung0112/hk_server/v5/shared/mlog"
	"github.com/masterhung0112/hk_server/v5/utils"
	"github.com/masterhung0112/hk_server/v5/web"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:          "server",
	Short:        "Run the server",
	RunE:         serverCmdF,
	SilenceUsage: false,
}

func init() {
	RootCmd.AddCommand(serverCmd)
	RootCmd.RunE = serverCmdF
}

func serverCmdF(command *cobra.Command, args []string) error {
	disableConfigWatch, _ := command.Flags().GetBool("disableconfigwatch")
	usedPlatform, _ := command.Flags().GetBool("platform")

	interruptChan := make(chan os.Signal, 1)

	if err := utils.TranslationsPreInit(); err != nil {
		return errors.Wrap(err, "unable to load Mattermost translation files")
	}

	customDefaults, err := loadCustomDefaults()
	if err != nil {
		mlog.Error("Error loading custom configuration defaults: " + err.Error())
	}

	configStore, err := config.NewStore(getConfigDSN(command, config.GetEnvironment()), !disableConfigWatch, customDefaults)
	if err != nil {
		return errors.Wrap(err, "failed to load configuration")
	}
	defer configStore.Close()

	return runServer(configStore, usedPlatform, interruptChan)
}

func runServer(configStore *config.Store, usedPlatform bool, interruptChan chan os.Signal) error {
	// Setting the highest traceback level from the code.
	// This is done to print goroutines from all threads (see golang.org/issue/13161)
	// and also preserve a crash dump for later investigation.
	debug.SetTraceback("crash")

	options := []app.Option{
		app.ConfigStore(configStore),
		app.RunJobs,
		app.JoinCluster,
		app.StartSearchEngine,
		app.StartMetrics,
	}
	server, err := app.NewServer(options...)
	if err != nil {
		mlog.Critical(err.Error())
		return err
	}
	defer server.Shutdown()

	if usedPlatform {
		mlog.Error("The platform binary has been deprecated, please switch to using the mattermost binary.")
	}

	api := api4.ApiInit(server, server.AppOptions, server.Router)
	//TODO: Open
	// wsapi.Init(server)
	web.New(server, server.AppOptions, server.Router)
	api4.ApiInitLocal(server, server.AppOptions, server.LocalRouter)

	serverErr := server.Start()
	if serverErr != nil {
		mlog.Critical(serverErr.Error())
		return serverErr
	}

	// If we allow testing then listen for manual testing URL hits
	if *server.Config().ServiceSettings.EnableTesting {
		manualtesting.Init(api)
	}

	notifyReady()

	// wait for kill signal before attempting to gracefully shutdown
	// the running service
	signal.Notify(interruptChan, syscall.SIGINT, syscall.SIGTERM)
	<-interruptChan

	return nil
}

func notifyReady() {
	// If the environment vars provide a systemd notification socket,
	// notify systemd that the server is ready.
	systemdSocket := os.Getenv("NOTIFY_SOCKET")
	if systemdSocket != "" {
		mlog.Info("Sending systemd READY notification.")

		err := sendSystemdReadyNotification(systemdSocket)
		if err != nil {
			mlog.Error(err.Error())
		}
	}
}

func sendSystemdReadyNotification(socketPath string) error {
	msg := "READY=1"
	addr := &net.UnixAddr{
		Name: socketPath,
		Net:  "unixgram",
	}
	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write([]byte(msg))
	return err
}
