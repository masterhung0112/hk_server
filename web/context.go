package web

import (
	"net/http"

	"github.com/masterhung0112/go_server/app"
	"github.com/masterhung0112/go_server/model"
)

type Context struct {
	App    *app.App
	Err    *model.AppError
	Params *Params
}

func NewInvalidParamError(parameter string) *model.AppError {
	err := model.NewAppError("Context", "api.context.invalid_body_param.app_error", map[string]interface{}{"Name": parameter}, "", http.StatusBadRequest)
	return err
}

func (c *Context) SetInvalidParam(parameter string) {
	c.Err = NewInvalidParamError(parameter)
}

func (c *Context) RequireUserId() *Context {
	if c.Err != nil {
		return c
	}

	if c.Params.UserId == model.ME {
		c.Params.UserId = c.App.Session().UserId
	}

	if !model.IsValidId(c.Params.UserId) {
		c.SetInvalidUrlParam("user_id")
	}
	return c
}

func (c *Context) SetInvalidUrlParam(parameter string) {
	c.Err = NewInvalidUrlParamError(parameter)
}

func NewInvalidUrlParamError(parameter string) *model.AppError {
	err := model.NewAppError("Context", "api.context.invalid_url_param.app_error", map[string]interface{}{"Name": parameter}, "", http.StatusBadRequest)
	return err
}

func (c *Context) SetPermissionError(permissions ...*model.Permission) {
	c.Err = c.App.MakePermissionError(permissions)
}

func (c *Context) IsSystemAdmin() bool {
	return c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM)
}

func (c *Context) HandleEtag(etag string, routeName string, w http.ResponseWriter, r *http.Request) bool {
	// metrics := c.App.Metrics()
	if et := r.Header.Get(model.HEADER_ETAG_CLIENT); len(etag) > 0 {
		if et == etag {
			w.Header().Set(model.HEADER_ETAG_SERVER, etag)
			w.WriteHeader(http.StatusNotModified)
			//TODO: Open
			// if metrics != nil {
			// 	metrics.IncrementEtagHitCounter(routeName)
			// }
			return true
		}
	}

	//TODO: Open this
	// if metrics != nil {
	// 	metrics.IncrementEtagMissCounter(routeName)
	// }

	return false
}
