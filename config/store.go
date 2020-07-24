package config

import (
  "github.com/masterhung0112/go_server/model"
)
// Listener is a callback function invoked when the configuration changes
type Listener func(oldConfig *model.Config, newConfig *model.Config)

type Store interface {
  // Get fetches the current, cached configuration.
  Get() *model.Config

  // Set replaces the current configuration in its entirety and updates the backing store
  Set(*model.Config) (*model.Config, error)

  AddListener(listener Listener) string
  RemoveListener(id string)
}