package web

import (
	"github.com/masterhung0112/go_server/services/configservice"
	"github.com/masterhung0112/go_server/utils"
	"net/http"
	"path"
	"strings"
)

func IsApiCall(config configservice.ConfigService, r *http.Request) bool {
	subpath, _ := utils.GetSubpathFromConfig(config.Config())

	return strings.HasPrefix(r.URL.Path, path.Join(subpath, "api")+"/")
}
