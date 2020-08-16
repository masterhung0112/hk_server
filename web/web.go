package web

import (
	"github.com/masterhung0112/go_server/services/configservice"
	"net/http"
	"path"
	"strings"
	"github.com/masterhung0112/go_server/utils"
)

func IsApiCall(config configservice.ConfigService, r *http.Request) bool {
	subpath, _ := utils.GetSubpathFromConfig(config.Config())

	return strings.HasPrefix(r.URL.Path, path.Join(subpath, "api")+"/")
}