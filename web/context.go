package web

import (
	"github.com/masterhung0112/hk_server/audit"
	"strings"
	"github.com/masterhung0112/hk_server/utils"
	"net/http"

	"github.com/masterhung0112/hk_server/app"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
)

type Context struct {
	App           *app.App
	Logger        *mlog.Logger
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


// LogAuditRec logs an audit record using default LevelAPI.
func (c *Context) LogAuditRec(rec *audit.Record) {
	c.LogAuditRecWithLevel(rec, app.LevelAPI)
}

// LogAuditRec logs an audit record using specified Level.
// If the context is flagged with a permissions error then `level`
// is ignored and the audit record is emitted with `LevelPerms`.
func (c *Context) LogAuditRecWithLevel(rec *audit.Record, level mlog.LogLevel) {
	if rec == nil {
		return
	}
	if c.Err != nil {
		rec.AddMeta("err", c.Err.Id)
		rec.AddMeta("code", c.Err.StatusCode)
		if c.Err.Id == "api.context.permissions.app_error" {
			level = app.LevelPerms
		}
		rec.Fail()
	}
	c.App.Srv().Audit.LogRecord(level, *rec)
}

// MakeAuditRecord creates a audit record pre-populated with data from this context.
func (c *Context) MakeAuditRecord(event string, initialStatus string) *audit.Record {
	rec := &audit.Record{
		APIPath:   c.App.Path(),
		Event:     event,
		Status:    initialStatus,
		UserID:    c.App.Session().UserId,
		SessionID: c.App.Session().Id,
		Client:    c.App.UserAgent(),
		IPAddress: c.App.IpAddress(),
		Meta:      audit.Meta{audit.KeyClusterID: c.App.GetClusterId()},
	}
	rec.AddMetaTypeConverter(model.AuditModelTypeConv)

	return rec
}

func (c *Context) LogAudit(extraInfo string) {
	audit := &model.Audit{UserId: c.App.Session().UserId, IpAddress: c.App.IpAddress(), Action: c.App.Path(), ExtraInfo: extraInfo, SessionId: c.App.Session().Id}
	if err := c.App.Srv().Store.Audit().Save(audit); err != nil {
		appErr := model.NewAppError("LogAudit", "app.audit.save.saving.app_error", nil, err.Error(), http.StatusInternalServerError)
		c.LogErrorByCode(appErr)
	}
}

func (c *Context) LogAuditWithUserId(userId, extraInfo string) {

	if len(c.App.Session().UserId) > 0 {
		extraInfo = strings.TrimSpace(extraInfo + " session_user=" + c.App.Session().UserId)
	}

	audit := &model.Audit{UserId: userId, IpAddress: c.App.IpAddress(), Action: c.App.Path(), ExtraInfo: extraInfo, SessionId: c.App.Session().Id}
	if err := c.App.Srv().Store.Audit().Save(audit); err != nil {
		appErr := model.NewAppError("LogAuditWithUserId", "app.audit.save.saving.app_error", nil, err.Error(), http.StatusInternalServerError)
		c.LogErrorByCode(appErr)
	}
}

func (c *Context) LogErrorByCode(err *model.AppError) {
	code := err.StatusCode
	msg := err.SystemMessage(utils.TDefault)
	fields := []mlog.Field{
		mlog.String("err_where", err.Where),
		mlog.Int("http_code", err.StatusCode),
		mlog.String("err_details", err.DetailedError),
	}
	switch {
	case (code >= http.StatusBadRequest && code < http.StatusInternalServerError) ||
		err.Id == "web.check_browser_compatibility.app_error":
		c.Logger.Debug(msg, fields...)
	case code == http.StatusNotImplemented:
		c.Logger.Info(msg, fields...)
	default:
		c.Logger.Error(msg, fields...)
	}
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
