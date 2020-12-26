package web

import (
	"github.com/masterhung0112/hk_server/utils"
	"net/http"

	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
)

type Context struct {
	App           *app.App
	Log           *mlog.Logger
	Err           *model.AppError
	Params        *Params
	siteURLHeader string
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

func (c *Context) RemoveSessionCookie(w http.ResponseWriter, r *http.Request) {
	subpath, _ := utils.GetSubpathFromConfig(c.App.Config())

	cookie := &http.Cookie{
		Name:     model.SESSION_COOKIE_TOKEN,
		Value:    "",
		Path:     subpath,
		MaxAge:   -1,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

func (c *Context) SessionRequired() {
	if !*c.App.Config().ServiceSettings.EnableUserAccessTokens &&
		c.App.Session().Props[model.SESSION_PROP_TYPE] == model.SESSION_TYPE_USER_ACCESS_TOKEN &&
		c.App.Session().Props[model.SESSION_PROP_IS_BOT] != model.SESSION_PROP_IS_BOT_VALUE {

		c.Err = model.NewAppError("", "api.context.session_expired.app_error", nil, "UserAccessToken", http.StatusUnauthorized)
		return
	}

	if len(c.App.Session().UserId) == 0 {
		c.Err = model.NewAppError("", "api.context.session_expired.app_error", nil, "UserRequired", http.StatusUnauthorized)
		return
	}
}
