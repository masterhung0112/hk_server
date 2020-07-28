package commands

import (
	"github.com/stretchr/testify/require"
	"github.com/masterhung0112/go_server/config"
	"github.com/masterhung0112/go_server/jobs"
	"syscall"
	"testing"
	"os"

)

const (
  UnitTestListeningPort = ":0"
)

type ServerTestHelper struct {
  disableConfigWatch bool
  interruptChan chan os.Signal
  originalInterval int
}

func SetupServerTest(t testing.TB) *ServerTestHelper {
  if testing.Short() {
    t.SkipNow()
  }

  // Build a channel that will be used by the server to receive system signals...
  interruptChan := make(chan os.Signal, 1)
  // ...and send itt immediate SIGINT value.
  // This will make server loop stop as soon as it started successfully.
  interruptChan <- syscall.SIGINT

  // Let jobs poll for termination every 0.2s (instead of every 15s by default)
	// Otherwise we would have to wait the whole polling duration before the test
  // terminates.

  originalInterval := jobs.DEFAULT_WATCHER_POLLING_INTERVAL
	jobs.DEFAULT_WATCHER_POLLING_INTERVAL = 200

	th := &ServerTestHelper{
		disableConfigWatch: true,
		interruptChan:      interruptChan,
		originalInterval:   originalInterval,
	}
	return th
}

func (th *ServerTestHelper) TearDownServerTest() {
	jobs.DEFAULT_WATCHER_POLLING_INTERVAL = th.originalInterval
}

func TestRunServerSuccess(t *testing.T) {
	th := SetupServerTest(t)
	defer th.TearDownServerTest()

	configStore, err := config.NewMemoryStore()
	require.NoError(t, err)

	// Use non-default listening port in case another server instance is already running.
	*configStore.Get().ServiceSettings.ListenAddress = UnitTestListeningPort

	err = runServer(configStore, th.interruptChan)//, th.disableConfigWatch, false
	require.NoError(t, err)
}