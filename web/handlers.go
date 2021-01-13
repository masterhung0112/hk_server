package web

import (
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/utils"
	"net/http"
	"reflect"
	"runtime"
	"strings"
)

type Handler struct {
	GetGlobalAppOptions app.AppOptionCreator
	HandleFunc          func(*Context, http.ResponseWriter, *http.Request)
	HandlerName         string
	RequireSession      bool
}

func GetHandlerName(h func(*Context, http.ResponseWriter, *http.Request)) string {
	handlerName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
	pos := strings.LastIndex(handlerName, ".")
	if pos != -1 && len(handlerName) > pos {
		handlerName = handlerName[pos+1:]
	}
	return handlerName
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Generate new request ID

	// Create new context
	c := &Context{}
	c.App = app.New(
		h.GetGlobalAppOptions()...,
	)
	c.App.InitServer()

	t, _ := utils.GetTranslationsAndLocale(w, r)
	c.App.SetT(t)
	c.Logger = c.App.Log()

	c.Params = ParamsFromRequest(r)

	token, tokenLocation := app.ParseAuthTokenFromRequest(r)

	if len(token) != 0 {
		session, err := c.App.GetSession(token)
		if err != nil {
			c.Logger.Info("Invalid session", mlog.Err(err))
			if err.StatusCode == http.StatusInternalServerError {
				c.Err = err
			} else if h.RequireSession {
				c.RemoveSessionCookie(w, r)
				c.Err = model.NewAppError("ServeHTTP", "api.context.session_expired.app_error", nil, "token="+token, http.StatusUnauthorized)
			}
		} else if !session.IsOAuth && tokenLocation == app.TokenLocationQueryString {
			c.Err = model.NewAppError("ServeHTTP", "api.context.token_provided.app_error", nil, "token="+token, http.StatusUnauthorized)
		} else {
			c.App.SetSession(session)
		}

		// Rate limit by UserID
		//TODO: Open
		// if c.App.Srv().RateLimiter != nil && c.App.Srv().RateLimiter.UserIdRateLimit(c.App.Session().UserId, w) {
		// 	return
		// }

		//TODO: Check
		// h.checkCSRFToken(c, r, token, tokenLocation, session)
	}

	if c.Err == nil && h.RequireSession {
		c.SessionRequired()
	}

	if c.Err == nil {
		// call the real handler
		h.HandleFunc(c, w, r)
	}

	// Handle errors that have occurred
	if c.Err != nil {
		// c.Err.Translate(c.App.T)
		// c.Err.RequestId = c.App.RequestId()

		if c.Err.Id == "api.context.session_expired.app_error" {
			// c.LogInfo(c.Err)
		} else {
			// c.LogError(c.Err)
		}

		c.Err.Where = r.URL.Path

		// Block out detailed error when not in developer mode
		if !*c.App.Config().ServiceSettings.EnableDeveloper {
			c.Err.DetailedError = ""
		}

		// Sanitize all 5xx error messages in hardened mode
		// if *c.App.Config().ServiceSettings.ExperimentalEnableHardenedMode && c.Err.StatusCode >= 500 {
		// 	c.Err.Id = ""
		// 	c.Err.Message = "Internal Server Error"
		// 	c.Err.DetailedError = ""
		// 	c.Err.StatusCode = 500
		// 	c.Err.Where = ""
		// 	// c.Err.IsOAuth = false
		// }

		// if IsApiCall(c.App, r) || IsWebhookCall(c.App, r) || IsOAuthApiCall(c.App, r) || len(r.Header.Get("X-Mobile-App")) > 0 {
		if IsApiCall(c.App, r) {
			w.WriteHeader(c.Err.StatusCode)
			w.Write([]byte(c.Err.ToJson()))
		} else {
			utils.RenderWebAppError(c.App.Config(), w, r, c.Err, nil) //c.App.AsymmetricSigningKey())
		}

		// if c.App.Metrics() != nil {
		// 	c.App.Metrics().IncrementHttpError()
		// }
	}
}
