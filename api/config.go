package api

import (
	"fmt"
	"github.com/masterhung0112/hk_server/config"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/utils"
	"net/http"
	"reflect"
	"strings"
)

var writeFilter func(c *Context, structField reflect.StructField) bool
var readFilter func(c *Context, structField reflect.StructField) bool
var permissionMap map[string]*model.Permission

type filterType string

const filterTypeWrite filterType = "write"
const filterTypeRead filterType = "read"

func (api *API) InitConfig() {
	api.BaseRoutes.ApiRoot.Handle("/config", api.ApiSessionRequired(getConfig)).Methods("GET")
	api.BaseRoutes.ApiRoot.Handle("/config/client", api.ApiHandler(getClientConfig)).Methods("GET")
}

func init() {
	writeFilter = makeFilterConfigByPermission(filterTypeWrite)
	readFilter = makeFilterConfigByPermission(filterTypeRead)
	permissionMap = map[string]*model.Permission{}
	for _, p := range model.AllPermissions {
		permissionMap[p.Id] = p
	}
}

func getConfig(c *Context, w http.ResponseWriter, r *http.Request) {
	if !c.App.SessionHasPermissionToAny(*c.App.Session(), model.SysconsoleReadPermissions) {
		c.SetPermissionError(model.SysconsoleReadPermissions...)
		return
	}

	// auditRec := c.MakeAuditRecord("getConfig", audit.Fail)
	// defer c.LogAuditRec(auditRec)

	cfg, err := config.Merge(&model.Config{}, c.App.GetSanitizedConfig(), &utils.MergeConfig{
		StructFieldFilter: func(structField reflect.StructField, base, patch reflect.Value) bool {
			return readFilter(c, structField)
		},
	})

	if err != nil {
		c.Err = model.NewAppError("getConfig", "api.config.get_config.restricted_merge.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	// auditRec.Success()

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Write([]byte(cfg.ToJson()))
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

func makeFilterConfigByPermission(accessType filterType) func(c *Context, structField reflect.StructField) bool {
	return func(c *Context, structField reflect.StructField) bool {
		if structField.Type.Kind() == reflect.Struct {
			return true
		}

		tagPermissions := strings.Split(structField.Tag.Get("access"), ",")

		// If there are no access tag values and the role has manage_system, no need to continue
		// checking permissions.
		if len(tagPermissions) == 0 {
			if c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM) {
				return true
			}
		}

		// one iteration for write_restrictable value, it could be anywhere in the order of values
		for _, val := range tagPermissions {
			tagValue := strings.TrimSpace(val)
			if tagValue == "" {
				continue
			}
			// ConfigAccessTagWriteRestrictable trumps all other permissions
			if tagValue == model.ConfigAccessTagWriteRestrictable {
				if *c.App.Config().ExperimentalSettings.RestrictSystemAdmin && accessType == filterTypeWrite {
					return false
				}
				continue
			}
		}

		// another iteration for permissions checks of other tag values
		for _, val := range tagPermissions {
			tagValue := strings.TrimSpace(val)
			if tagValue == "" {
				continue
			}
			if tagValue == model.ConfigAccessTagWriteRestrictable {
				continue
			}
			permissionID := fmt.Sprintf("sysconsole_%s_%s", accessType, tagValue)
			if permission, ok := permissionMap[permissionID]; ok {
				if c.App.SessionHasPermissionTo(*c.App.Session(), permission) {
					return true
				}
			} else {
				mlog.Warn("Unrecognized config permissions tag value.", mlog.String("tag_value", permissionID))
			}
		}

		// with manage_system, default to allow, otherwise default not-allow
		return c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_MANAGE_SYSTEM)
	}
}
