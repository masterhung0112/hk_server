package config

import (
	"bytes"
	"github.com/masterhung0112/go_server/model"
	"github.com/pkg/errors"
	"io/ioutil"
)

type MemoryStore struct {
	commonStore

	allowEnvironmentOverrides bool
	validate                  bool
	files                     map[string][]byte
	savedConfig               *model.Config
}

// MemoryStoreOptions makes configuration of the memory store explicit.
type MemoryStoreOptions struct {
	IgnoreEnvironmentOverrides bool
	SkipValidation             bool
	InitialConfig              *model.Config
	InitialFiles               map[string][]byte
}

// NewMemoryStore creates a new MemoryStore instance with default options.
func NewMemoryStore() (*MemoryStore, error) {
	return NewMemoryStoreWithOptions(&MemoryStoreOptions{})
}

// NewMemoryStoreWithOptions creates a new MemoryStore instance.
func NewMemoryStoreWithOptions(options *MemoryStoreOptions) (*MemoryStore, error) {
	savedConfig := options.InitialConfig
	if savedConfig == nil {
		savedConfig = &model.Config{}
		savedConfig.SetDefaults()
	}

	initialFiles := options.InitialFiles
	if initialFiles == nil {
		initialFiles = make(map[string][]byte)
	}

	ms := &MemoryStore{
		allowEnvironmentOverrides: !options.IgnoreEnvironmentOverrides,
		validate:                  !options.SkipValidation,
		files:                     initialFiles,
		savedConfig:               savedConfig,
	}

	ms.commonStore.config = &model.Config{}
	ms.commonStore.config.SetDefaults()

	if err := ms.Load(); err != nil {
		return nil, err
	}

	return ms, nil
}

// Set replaces the current configuration in its entirety.
func (ms *MemoryStore) Set(newCfg *model.Config) (*model.Config, error) {
	validate := ms.commonStore.validate
	if !ms.validate {
		validate = nil
	}

	return ms.commonStore.set(newCfg, ms.allowEnvironmentOverrides, validate, ms.persist)
}

// persist copies the active config to the saved config.
func (ms *MemoryStore) persist(cfg *model.Config) error {
	ms.savedConfig = cfg.Clone()

	return nil
}

// Load applies environment overrides to the default config as if a re-load had occurred.
func (ms *MemoryStore) Load() (err error) {
	var cfgBytes []byte
	cfgBytes, err = marshalConfig(ms.savedConfig)
	if err != nil {
		return errors.Wrap(err, "failed to serialize config")
	}

	f := ioutil.NopCloser(bytes.NewReader(cfgBytes))

	validate := ms.commonStore.validate
	if !ms.validate {
		validate = nil
	}

	return ms.commonStore.load(f, false, validate, ms.persist)
}
