package api

import (
	"net/http"

	"github.com/masterhung0112/go_server/web"
)

// ApiSessionRequired provides a handler for API endpoints which require the user to be logged in in order for access to
// be granted.
func (api *API) ApiSessionRequired(h func(*web.Context, http.ResponseWriter, *http.Request)) http.Handler {
	handler := &web.Handler{
		HandleFunc:  h,
		HandlerName: web.GetHandlerName(h),
	}

	return handler
}
