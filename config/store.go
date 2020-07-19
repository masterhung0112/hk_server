package config

import (
  "github.com/masterhung0112/go_server/model"
)

type Store interface {
  // Get fetches the current, cached configuration.
  Get() *model.Config
}