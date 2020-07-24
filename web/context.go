package web

import (
	"github.com/masterhung0112/go_server/model"
  "github.com/masterhung0112/go_server/app"
)

type Context struct {
  App           *app.App
  Err           *model.AppError
}
