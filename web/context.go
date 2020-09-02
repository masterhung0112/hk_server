package web

import (
  "net/http"

	"github.com/masterhung0112/go_server/model"
  "github.com/masterhung0112/go_server/app"
)

type Context struct {
  App           *app.App
  Err           *model.AppError
}

func NewInvalidParamError(parameter string) *model.AppError {
	err := model.NewAppError("Context", "api.context.invalid_body_param.app_error", map[string]interface{}{"Name": parameter}, "", http.StatusBadRequest)
	return err
}

func (c *Context) SetInvalidParam(parameter string) {
	c.Err = NewInvalidParamError(parameter)
}