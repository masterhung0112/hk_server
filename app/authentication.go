/**
 * Contains a list of functions related to authentication
 */
package app

import (
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/utils"
)

func (app *App) IsPasswordValid(password string) *model.AppError {
	return utils.IsPasswordValidWithSettings(password, &app.Config().PasswordSettings)
}
