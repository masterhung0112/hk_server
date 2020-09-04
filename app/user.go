/**
 * Contains a list of functions related to user
 * - User creation
 * - Soft-delete user
 */

package app

import (
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
)

func (s *Server) IsFirstUserAccount() bool {
	// cachedSessions, err := s.sessionCache.Len()
	// if err != nil {
	// 	return false
	// }
	// if cachedSessions == 0 {
		count, err := s.Store.User().Count(model.UserCountOptions{IncludeDeleted: true})
		if err != nil {
			mlog.Error("There was a error fetching if first user account", mlog.Err(err))
			return false
		}
		if count <= 0 {
			return true
		}
	// }

	return false
}

/****************
 * User Creation
 ****************/

// CreateGuest creates a guest and sets several fields of the returned User struct to
// their zero values.
func (a *App) CreateGuest(user *model.User) (*model.User, *model.AppError) {
	return a.createUserOrGuest(user, true)
}

// CreateUser creates a user and sets several fields of the returned User struct to
// their zero values.
func (app *App) CreateUser(user *model.User) (*model.User, *model.AppError) {
	return app.createUserOrGuest(user, false)
}

// Create new User from QR Code or something
func (app *App) CreateUserWithToken(user *model.User, token *model.Token) (*model.User, *model.AppError) {
	// Check if token is expired or not

	// Create tokenData from Extra information

	// Get team from token extra

	// Get channel from token extra

	// Get email from token extra

	// Create New User

	// Add User To Team

	// Add
	return nil, nil
}

func (app *App) createUserOrGuest(user *model.User, guest bool) (*model.User, *model.AppError) {
	// Define a list of default role
	user.Roles = model.SYSTEM_USER_ROLE_ID
	if guest {
		user.Roles = model.SYSTEM_GUEST_ROLE_ID
	}

	// if !user.IsLDAPUser() && !user.IsSAMLUser() && !user.IsGuest() && !CheckUserDomain(user, *a.Config().TeamSettings.RestrictCreationToDomains) {
	// 	return nil, model.NewAppError("CreateUser", "api.user.create_user.accepted_domain.app_error", nil, "", http.StatusBadRequest)
	// }

	// if !user.IsLDAPUser() && !user.IsSAMLUser() && user.IsGuest() && !CheckUserDomain(user, *a.Config().GuestAccountsSettings.RestrictCreationToDomains) {
	// 	return nil, model.NewAppError("CreateUser", "api.user.create_user.accepted_domain.app_error", nil, "", http.StatusBadRequest)
	// }

	//TODO: Count the number of user, if we have never created user before, that is the admin user
	count, err := app.Srv().Store.User().Count(model.UserCountOptions{IncludeDeleted: true})
	if err != nil {
		return nil, err
  }
  if count <= 0 {
		user.Roles = model.SYSTEM_ADMIN_ROLE_ID + " " + model.SYSTEM_USER_ROLE_ID
  }

  // if _, ok := utils.GetSupportedLocales()[user.Locale]; !ok {
	// 	user.Locale = *a.Config().LocalizationSettings.DefaultClientLocale
	// }

	ruser, err := app.createUser(user)
	if err != nil {
		return nil, err
	}

	// Check is password valid

	// Save the user into DB

	// Mark email as verified if necessary

	return ruser, nil
}

func (a *App) createUser(user *model.User) (*model.User, *model.AppError) {
  user.MakeNonNil()

  // if err := a.IsPasswordValid(user.Password); user.AuthService == "" && err != nil {
  if err := a.IsPasswordValid(user.Password); err != nil {
		return nil, err
  }

  ruser, err := a.Srv().Store.User().Save(user)
	if err != nil {
		mlog.Error("Couldn't save the user", mlog.Err(err))
		return nil, err
  }

  if user.EmailVerified {
		if err := a.VerifyUserEmail(ruser.Id, user.Email); err != nil {
			mlog.Error("Failed to set email verified", mlog.Err(err))
		}
	}

	// pref := model.Preference{UserId: ruser.Id, Category: model.PREFERENCE_CATEGORY_TUTORIAL_STEPS, Name: ruser.Id, Value: "0"}
	// if err := a.Srv().Store.Preference().Save(&model.Preferences{pref}); err != nil {
	// 	mlog.Error("Encountered error saving tutorial preference", mlog.Err(err))
	// }

	ruser.Sanitize(map[string]bool{})
	return ruser, nil
}

/*
 * Create user through sign up screen
 */
func (app *App) CreateUserFromSignup(user *model.User) (*model.User, *model.AppError) {
	//TODO: Prevent signup if setting doesn't allow it

	//TODO: if this user is the first account, create with admin role

	user.EmailVerified = false

	ruser, err := app.CreateUser(user)
	if err != nil {
		return nil, err
	}

	//TODO: Send email if necessary
	// if err := a.Srv().EmailService.sendWelcomeEmail(ruser.Id, ruser.Email, ruser.EmailVerified, ruser.Locale, a.GetSiteURL(), redirect); err != nil {
	// mlog.Error("Failed to send welcome email on create user from signup", mlog.Err(err))
	// }

	return ruser, nil
}

func (a *App) VerifyUserEmail(userId, email string) *model.AppError {
  //TODO: Open
	// if _, err := a.Srv().Store.User().VerifyEmail(userId, email); err != nil {
	// 	return err
	// }

	// a.InvalidateCacheForUser(userId)

	// user, err := a.GetUser(userId)

	// if err != nil {
	// 	return err
	// }

	// a.sendUpdatedUserEvent(*user)

	return nil
}

/****************
 * Login
 ****************/
// func (app *App) Login(login_id string, password string) {
//
// }

/****************
 * User status
 ****************/

func (a *App) IsFirstUserAccount() bool {
	return a.Srv().IsFirstUserAccount()
}
