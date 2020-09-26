package web

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/services/configservice"
	"github.com/masterhung0112/hk_server/utils"
	"net/http"
	"path"
	"strings"
)

func IsApiCall(config configservice.ConfigService, r *http.Request) bool {
	subpath, _ := utils.GetSubpathFromConfig(config.Config())

	return strings.HasPrefix(r.URL.Path, path.Join(subpath, "api")+"/")
}

func ReturnStatusOK(w http.ResponseWriter) {
	m := make(map[string]string)
	m[model.STATUS] = model.STATUS_OK
	w.Write([]byte(model.MapToJson(m)))
}
