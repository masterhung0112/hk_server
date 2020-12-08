package sqlstore

import (
	"github.com/masterhung0112/hk_server/model"
)

func MakeEmail() string {
	return "success_" + model.NewId() + "@simulator.amazonses.com"
}
