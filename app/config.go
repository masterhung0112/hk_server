package app

import (
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/utils"
	"net/url"
	"strconv"
	"time"
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

func (a *App) ClientConfig() map[string]string {
	return a.Srv().clientConfig.Load().(map[string]string)
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

func (s *Server) PostActionCookieSecret() []byte {
	return s.postActionCookieSecret
}

func (a *App) PostActionCookieSecret() []byte {
	return a.Srv().PostActionCookieSecret()
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

func (s *Server) ensureInstallationDate() error {
	_, appErr := s.getSystemInstallDate()
	if appErr == nil {
		return nil
	}

	installDate, appErr := s.Store.User().InferSystemInstallDate()
	var installationDate int64
	if appErr == nil && installDate > 0 {
		installationDate = installDate
	} else {
		installationDate = utils.MillisFromTime(time.Now())
	}

	if err := s.Store.System().SaveOrUpdate(&model.System{
		Name:  model.SYSTEM_INSTALLATION_DATE_KEY,
		Value: strconv.FormatInt(installationDate, 10),
	}); err != nil {
		return err
	}
	return nil
}

func (s *Server) regenerateClientConfig() {
	clientConfig := config.GenerateClientConfig(s.Config(), "", nil)               // s.diagnosticId, s.License())
	limitedClientConfig := config.GenerateLimitedClientConfig(s.Config(), "", nil) //s.diagnosticId, s.License())

	//TODO: Open this
	// if clientConfig["EnableCustomTermsOfService"] == "true" {
	// 	termsOfService, err := s.Store.TermsOfService().GetLatest(true)
	// 	if err != nil {
	// 		mlog.Err(err)
	// 	} else {
	// 		clientConfig["CustomTermsOfServiceId"] = termsOfService.Id
	// 		limitedClientConfig["CustomTermsOfServiceId"] = termsOfService.Id
	// 	}
	// }

	if key := s.AsymmetricSigningKey(); key != nil {
		//TODO: Open
		// der, _ := x509.MarshalPKIXPublicKey(&key.PublicKey)
		// clientConfig["AsymmetricSigningPublicKey"] = base64.StdEncoding.EncodeToString(der)
		// limitedClientConfig["AsymmetricSigningPublicKey"] = base64.StdEncoding.EncodeToString(der)
	}

	clientConfigJSON, _ := json.Marshal(clientConfig)
	s.clientConfig.Store(clientConfig)
	s.limitedClientConfig.Store(limitedClientConfig)
	s.clientConfigHash.Store(fmt.Sprintf("%x", md5.Sum(clientConfigJSON)))
}

func (a *App) GetCookieDomain() string {
	if *a.Config().ServiceSettings.AllowCookiesForSubdomains {
		if siteURL, err := url.Parse(*a.Config().ServiceSettings.SiteURL); err == nil {
			return siteURL.Hostname()
		}
	}
	return ""
}

// GetSanitizedConfig gets the configuration for a system admin without any secrets.
func (a *App) GetSanitizedConfig() *model.Config {
	cfg := a.Config().Clone()
	cfg.Sanitize()

	return cfg
}
