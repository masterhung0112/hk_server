package app

import (
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"github.com/pkg/errors"
	"net/http"
)

func (a *App) CreateSession(session *model.Session) (*model.Session, *model.AppError) {
	session.Token = ""

	session, err := a.Srv().Store.Session().Save(session)
	if err != nil {
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &invErr):
			return nil, model.NewAppError("CreateSession", "app.session.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("CreateSession", "app.session.save.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

  //TODO: Open
	// a.AddSessionToCache(session)

	return session, nil
}

func (a *App) RevokeSessionById(sessionId string) *model.AppError {
	session, err := a.Srv().Store.Session().Get(sessionId)
	if err != nil {
		return model.NewAppError("RevokeSessionById", "app.session.get.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	return a.RevokeSession(session)
}

func (a *App) RevokeSession(session *model.Session) *model.AppError {
	if session.IsOAuth {
		if err := a.RevokeAccessToken(session.Token); err != nil {
			return err
		}
	} else {
		if err := a.Srv().Store.Session().Remove(session.Id); err != nil {
			return model.NewAppError("RevokeSession", "app.session.remove.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	//TODO: Open
	// a.ClearSessionCacheForUser(session.UserId)

	return nil
}

func (a *App) createSessionForUserAccessToken(tokenString string) (*model.Session, *model.AppError) {
	token, nErr := a.Srv().Store.UserAccessToken().GetByToken(tokenString)
	if nErr != nil {
		return nil, model.NewAppError("createSessionForUserAccessToken", "app.user_access_token.invalid_or_missing", nil, nErr.Error(), http.StatusUnauthorized)
	}

	if !token.IsActive {
		return nil, model.NewAppError("createSessionForUserAccessToken", "app.user_access_token.invalid_or_missing", nil, "inactive_token", http.StatusUnauthorized)
	}

	user, err := a.Srv().Store.User().Get(token.UserId)
	if err != nil {
		return nil, err
	}

	//TODO: Open
	//if !*a.Config().ServiceSettings.EnableUserAccessTokens && !user.IsBot {
	if !*a.Config().ServiceSettings.EnableUserAccessTokens {
		return nil, model.NewAppError("createSessionForUserAccessToken", "app.user_access_token.invalid_or_missing", nil, "EnableUserAccessTokens=false", http.StatusUnauthorized)
	}

	if user.DeleteAt != 0 {
		return nil, model.NewAppError("createSessionForUserAccessToken", "app.user_access_token.invalid_or_missing", nil, "inactive_user_id="+user.Id, http.StatusUnauthorized)
	}

	session := &model.Session{
		Token:   token.Token,
		UserId:  user.Id,
		Roles:   user.GetRawRoles(),
		IsOAuth: false,
	}

	session.AddProp(model.SESSION_PROP_USER_ACCESS_TOKEN_ID, token.Id)
	session.AddProp(model.SESSION_PROP_TYPE, model.SESSION_TYPE_USER_ACCESS_TOKEN)
	//TODO: Open
	// if user.IsBot {
	// 	session.AddProp(model.SESSION_PROP_IS_BOT, model.SESSION_PROP_IS_BOT_VALUE)
	// }
	if user.IsGuest() {
		session.AddProp(model.SESSION_PROP_IS_GUEST, "true")
	} else {
		session.AddProp(model.SESSION_PROP_IS_GUEST, "false")
	}
	a.SetSessionExpireInDays(session, model.SESSION_USER_ACCESS_TOKEN_EXPIRY)

	session, nErr = a.Srv().Store.Session().Save(session)
	if nErr != nil {
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(nErr, &invErr):
			return nil, model.NewAppError("CreateSession", "app.session.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("CreateSession", "app.session.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	a.AddSessionToCache(session)

	return session, nil
}

func (a *App) GetSession(token string) (*model.Session, *model.AppError) {
	//TODO: Open
	// metrics := a.Metrics()

	var session *model.Session
	var err *model.AppError
	//TODO: Open
	// if err := a.Srv().sessionCache.Get(token, &sess?ion); err == nil {
	// if metrics != nil {
	// 	metrics.IncrementMemCacheHitCounterSession()
	// }
	// } else {
	// if metrics != nil {
	// 	metrics.IncrementMemCacheMissCounterSession()
	// }
	// }

	if session == nil {
		var nErr error
		if session, nErr = a.Srv().Store.Session().Get(token); nErr == nil {
			if session != nil {
				if session.Token != token {
					return nil, model.NewAppError("GetSession", "api.context.invalid_token.error", map[string]interface{}{"Token": token, "Error": ""}, "session token is different from the one in DB", http.StatusUnauthorized)
				}

				if !session.IsExpired() {
					a.AddSessionToCache(session)
				}
			}
		} else if nfErr := new(store.ErrNotFound); !errors.As(nErr, &nfErr) {
			return nil, model.NewAppError("GetSession", "app.session.get.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
	}

	if session == nil {
		session, err = a.createSessionForUserAccessToken(token)
		if err != nil {
			detailedError := ""
			statusCode := http.StatusUnauthorized
			if err.Id != "app.user_access_token.invalid_or_missing" {
				detailedError = err.Error()
				statusCode = err.StatusCode
			} else {
				mlog.Warn("Error while creating session for user access token", mlog.Err(err))
			}
			return nil, model.NewAppError("GetSession", "api.context.invalid_token.error", map[string]interface{}{"Token": token, "Error": detailedError}, "", statusCode)
		}
	}

	if session == nil || session.IsExpired() {
		return nil, model.NewAppError("GetSession", "api.context.invalid_token.error", map[string]interface{}{"Token": token, "Error": ""}, "session is either nil or expired", http.StatusUnauthorized)
	}

	if *a.Config().ServiceSettings.SessionIdleTimeoutInMinutes > 0 &&
		!session.IsOAuth && !session.IsMobileApp() &&
		session.Props[model.SESSION_PROP_TYPE] != model.SESSION_TYPE_USER_ACCESS_TOKEN &&
		!*a.Config().ServiceSettings.ExtendSessionLengthWithActivity {

		timeout := int64(*a.Config().ServiceSettings.SessionIdleTimeoutInMinutes) * 1000 * 60
		if (model.GetMillis() - session.LastActivityAt) > timeout {
			// Revoking the session is an asynchronous task anyways since we are not checking
			// for the return value of the call before returning the error.
			// So moving this to a goroutine has 2 advantages:
			// 1. We are treating this as a proper asynchronous task.
			// 2. This also fixes a race condition in the web hub, where GetSession
			// gets called from (*WebConn).isMemberOfTeam and revoking a session involves
			// clearing the webconn cache, which needs the hub again.
			a.Srv().Go(func() {
				err := a.RevokeSessionById(session.Id)
				if err != nil {
					mlog.Warn("Error while revoking session", mlog.Err(err))
				}
			})
			return nil, model.NewAppError("GetSession", "api.context.invalid_token.error", map[string]interface{}{"Token": token, "Error": ""}, "idle timeout", http.StatusUnauthorized)
		}
	}

	return session, nil
}

// SetSessionExpireInDays sets the session's expiry the specified number of days
// relative to either the session creation date or the current time, depending
// on the `ExtendSessionOnActivity` config setting.
func (a *App) SetSessionExpireInDays(session *model.Session, days int) {
	if session.CreateAt == 0 || *a.Config().ServiceSettings.ExtendSessionLengthWithActivity {
		session.ExpiresAt = model.GetMillis() + (1000 * 60 * 60 * 24 * int64(days))
	} else {
		session.ExpiresAt = session.CreateAt + (1000 * 60 * 60 * 24 * int64(days))
	}
}

func (a *App) AddSessionToCache(session *model.Session) {
	//TODO: Open
	// a.Srv().sessionCache.SetWithExpiry(session.Token, session, time.Duration(int64(*a.Config().ServiceSettings.SessionCacheInMinutes))*time.Minute)
}

func (a *App) RevokeSessionsForDeviceId(userId string, deviceId string, currentSessionId string) *model.AppError {
	sessions, err := a.Srv().Store.Session().GetSessions(userId)
	if err != nil {
		return model.NewAppError("RevokeSessionsForDeviceId", "app.session.get_sessions.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	for _, session := range sessions {
		if session.DeviceId == deviceId && session.Id != currentSessionId {
			mlog.Debug("Revoking sessionId for userId. Re-login with the same device Id", mlog.String("session_id", session.Id), mlog.String("user_id", userId))
			if err := a.RevokeSession(session); err != nil {
				// Soft error so we still remove the other sessions
				mlog.Error(err.Error())
			}
		}
	}

	return nil
}
