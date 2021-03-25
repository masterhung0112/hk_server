package config

import (
	"encoding/json"

	"github.com/masterhung0112/hk_server/v5/model"
)

// marshalConfig converts the given configuration into JSON bytes for persistence.
func marshalConfig(cfg *model.Config) ([]byte, error) {
	return json.MarshalIndent(cfg, "", "    ")
}
