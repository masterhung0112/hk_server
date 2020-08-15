/**
 * Contains a list of functions related to user
 * - User creation
 * - Soft-delete user
 */

package app

import (
  "github.com/masterhung0112/go_server/model"
)

/****************
 * User Creation
 ****************/

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
  if (guest) {
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

  // Check is password valid

  // Save the user into DB

  // Mark email as verified if necessary

  return nil, nil
}

/*
 * Create user through sign up screen
 */
func (app *App) CreateUserFromSignup(user *model.User) (*model.User, *model.AppError){
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

/****************
 * Login
 ****************/
// func (app *App) Login(login_id string, password string) {
  //
// }