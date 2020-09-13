package app

import (
	"github.com/masterhung0112/go_server/model"
	"net/http"
)

func (a *App) RevokeAccessToken(token string) *model.AppError {
	// session, _ := a.GetSession(token)

	schan := make(chan error, 1)
	go func() {
		schan <- a.Srv().Store.Session().Remove(token)
		close(schan)
	}()

	//TODO: Open
	// if _, err := a.Srv().Store.OAuth().GetAccessData(token); err != nil {
	// 	return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.get.app_error", nil, "", http.StatusBadRequest)
	// }

	// if err := a.Srv().Store.OAuth().RemoveAccessData(token); err != nil {
	// 	return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.del_token.app_error", nil, "", http.StatusInternalServerError)
	// }

	if err := <-schan; err != nil {
		return model.NewAppError("RevokeAccessToken", "api.oauth.revoke_access_token.del_session.app_error", nil, "", http.StatusInternalServerError)
	}

	//TODO: Open
	// if session != nil {
	// 	a.ClearSessionCacheForUser(session.UserId)
	// }

	return nil
}
