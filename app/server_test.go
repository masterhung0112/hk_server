package app

import (
	"github.com/masterhung0112/go_server/model"
	"strconv"
	"net/http"
	"github.com/stretchr/testify/require"
	"testing"

)

func TestStartServerSuccess(t *testing.T) {
  s, err := NewServer()
  require.NoError(t, err)

  s.UpdateConfig(func(cfg *model.Config) {
    *cfg.ServiceSettings.ListenAddress = ":0"
  })
  serverErr := s.Start()
  require.NoError(t, serverErr)

  client := &http.Client{}
  checkEndpoint(t, client, "http://localhost:"+strconv.Itoa(s.ListenAddr.Port)+"/", http.StatusNotFound)

  s.Shutdown()
}

func checkEndpoint(t *testing.T, client *http.Client, url string, expectedStatus int) error {
  res, err := client.Get(url)

  if err != nil {
    return err
  }

  defer res.Body.Close()

  if res.StatusCode != expectedStatus {
    t.Errorf("Response code was %d; want %d", res.StatusCode, expectedStatus)
  }

  return nil
}