package config

import (
	"strings"

	"github.com/masterhung0112/hk_server/model"
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

// NewStore creates a database or file store given a data source name by which to connect.
func NewStore(dsn string, watch bool) (Store, error) {
	if strings.HasPrefix(dsn, "mysql://") || strings.HasPrefix(dsn, "postgres://") {
		return NewDatabaseStore(dsn)
	}

	return NewFileStore(dsn, watch)
}
