package storetest

import (
	"github.com/masterhung0112/go_server/model"
)

func MakeEmail() string {
	return "success_" + model.NewId() + "@simulator.amazonses.com"
}