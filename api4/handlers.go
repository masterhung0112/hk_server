package api1

import (
	"net/http"

	"github.com/masterhung0112/hk_server/web"
)

type Context = web.Context

func (api *API) ApiHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	handler := &web.Handler{
		GetGlobalAppOptions: api.GetGlobalAppOptions,
		HandleFunc:          h,
		HandlerName:         web.GetHandlerName(h),
		RequireSession:      false,
		TrustRequester:      false,
		RequireMfa:          false,
		IsStatic:            false,
		IsLocal:             false,
	}
	// if *api.ConfigService.Config().ServiceSettings.WebserverMode == "gzip" {
	// 	return gziphandler.GzipHandler(handler)
	// }
	return handler
}

// DisableWhenBusy provides a handler for API endpoints which should be disabled when the server is under load,
// responding with HTTP 503 (Service Unavailable).
func (api *API) ApiSessionRequiredDisableWhenBusy(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	handler := &web.Handler{
		GetGlobalAppOptions: api.GetGlobalAppOptions,
		HandleFunc:          h,
		HandlerName:         web.GetHandlerName(h),
		RequireSession:      true,
		TrustRequester:      false,
		RequireMfa:          false,
		IsStatic:            false,
		IsLocal:             false,
		DisableWhenBusy:     true,
	}
	// if *api.ConfigService.Config().ServiceSettings.WebserverMode == "gzip" {
	// 	return gziphandler.GzipHandler(handler)
	// }
	return handler

}
