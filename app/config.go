package app

import (
	"strconv"
	"crypto/ecdsa"
	"github.com/masterhung0112/go_server/model"
)

func (s *Server) Config() *model.Config {
	return s.configStore.Get()
}

func (a *App) Config() *model.Config {
	return a.Srv().Config()
}

func (a *App) UpdateConfig(f func(*model.Config)) {
	a.Srv().UpdateConfig(f)
}

// Registers a function with a given listener to be called when the config is reloaded and may have changed. The function
// will be called with two arguments: the old config and the new config. AddConfigListener returns a unique ID
// for the listener that can later be used to remove it.
func (s *Server) AddConfigListener(listener func(*model.Config, *model.Config)) string {
	return s.configStore.AddListener(listener)
}

func (a *App) AddConfigListener(listener func(*model.Config, *model.Config)) string {
	return a.Srv().AddConfigListener(listener)
}

// Removes a listener function by the unique ID returned when AddConfigListener was called
func (s *Server) RemoveConfigListener(id string) {
	s.configStore.RemoveListener(id)
}

func (a *App) RemoveConfigListener(id string) {
	a.Srv().RemoveConfigListener(id)
}

// AsymmetricSigningKey will return a private key that can be used for asymmetric signing.
func (s *Server) AsymmetricSigningKey() *ecdsa.PrivateKey {
  //TODO: Open
	return nil //s.asymmetricSigningKey
}

func (a *App) AsymmetricSigningKey() *ecdsa.PrivateKey {
	return a.Srv().AsymmetricSigningKey()
}

// ClientConfigWithComputed gets the configuration in a format suitable for sending to the client.
func (s *Server) ClientConfigWithComputed() map[string]string {
	respCfg := map[string]string{}
	for k, v := range s.clientConfig.Load().(map[string]string) {
		respCfg[k] = v
	}

	// These properties are not configurable, but nevertheless represent configuration expected
	// by the client.
  respCfg["NoAccounts"] = strconv.FormatBool(s.IsFirstUserAccount())

  //TODO: Open this code
	// respCfg["MaxPostSize"] = strconv.Itoa(s.MaxPostSize())
	// respCfg["UpgradedFromTE"] = strconv.FormatBool(s.isUpgradedFromTE())
	// respCfg["InstallationDate"] = ""
	// if installationDate, err := s.getSystemInstallDate(); err == nil {
	// 	respCfg["InstallationDate"] = strconv.FormatInt(installationDate, 10)
	// }

	return respCfg
}

func (a *App) LimitedClientConfig() map[string]string {
	return a.Srv().limitedClientConfig.Load().(map[string]string)
}

// ClientConfigWithComputed gets the configuration in a format suitable for sending to the client.
func (a *App) ClientConfigWithComputed() map[string]string {
	return a.Srv().ClientConfigWithComputed()
}

// LimitedClientConfigWithComputed gets the configuration in a format suitable for sending to the client.
func (a *App) LimitedClientConfigWithComputed() map[string]string {
	respCfg := map[string]string{}
	for k, v := range a.LimitedClientConfig() {
		respCfg[k] = v
	}

	// These properties are not configurable, but nevertheless represent configuration expected
	// by the client.
	respCfg["NoAccounts"] = strconv.FormatBool(a.IsFirstUserAccount())

	return respCfg
}