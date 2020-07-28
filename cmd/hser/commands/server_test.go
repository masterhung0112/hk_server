package commands

import (
	"syscall"
	"testing"
	"os"

)

const (
  UnitTestListeningPort = ":0"
)

type ServerTestHelper struct {
  dispatchConfigWatch bool
  interuptChan chan os.Signal
  originalInterval int
}

func SetupServerTest(t testing.TB) *ServerTestHelper {
  if testing.Short() {
    t.SkipNow()
  }

  // Build a channel that will be used by the server to receive system signals...
  interuptChan := make(chan os.Signal, 1)
  // ...and send itt immediate SIGINT value.
  // This will make server loop stop as soon as it started successfully.
  interuptChan <- syscall.SIGINT

  // Let jobs poll for termination every 0.2s (instead of every 15s by default)
	// Otherwise we would have to wait the whole polling duration before the test
  // terminates.

  originalInterval := jobs.DEFAULT_WATCHER_POLLING_INTERVAL
	// jobs.DEFAULT_WATCHER_POLLING_INTERVAL = 200

	th := &ServerTestHelper{
		disableConfigWatch: true,
		interruptChan:      interruptChan,
		originalInterval:   originalInterval,
	}
	return th
}