package api

import (
	"github.com/masterhung0112/go_server/model"
	"net/http"
)

func (api *API) InitConfig() {

	api.BaseRoutes.ApiRoot.Handle("/config/client", api.ApiHandler(getClientConfig)).Methods("GET")

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
	respCfg["MaxPostSize"] = strconv.Itoa(s.MaxPostSize())
	respCfg["UpgradedFromTE"] = strconv.FormatBool(s.isUpgradedFromTE())
	respCfg["InstallationDate"] = ""
	if installationDate, err := s.getSystemInstallDate(); err == nil {
		respCfg["InstallationDate"] = strconv.FormatInt(installationDate, 10)
	}

	return respCfg
}

func getClientConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")

	if format == "" {
		c.Err = model.NewAppError("getClientConfig", "api.config.client.old_format.app_error", nil, "", http.StatusNotImplemented)
		return
	}

	if format != "old" {
		c.SetInvalidParam("format")
		return
	}

	var config map[string]string
	if len(c.App.Session().UserId) == 0 {
		config = c.App.LimitedClientConfigWithComputed()
	} else {
		config = c.App.ClientConfigWithComputed()
	}

	w.Write([]byte(model.MapToJson(config)))
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