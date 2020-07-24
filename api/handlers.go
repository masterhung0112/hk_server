package api

import (
	"github.com/masterhung0112/go_server/web"
  "net/http"
)

type Context = web.Context

func (api *API) ApiHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
  handler := &web.Handler{
      HandleFunc: h,
  }
  return handler
}