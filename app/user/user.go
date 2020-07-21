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
}

func (app *App) createUserOrGuest(user *model.User, guest bool) (*model.User, *model.AppError) {
  // Define a list of default role
  user.Roles = model.SYSTEM_USER_ROLE_ID

  //TODO: Count the number of user, if we have never created user before, that is the admin user


  // Check is password valid

  // Save the user into DB

  // Mark email as verified if necessary


}

/*
 * Create user through sign up screen
 */
func (app *App) CreateUserFromSignup(user *model.User) (*model.User, *model.AppError){
  //TODO: if this user is the first account, create with admin role

  ruser, err := app.createUserOrGuest(user)
  if err != nil {
    return nil, err
  }

  //TODO: Send email if necessary

  return ruser, nil
}

/****************
 * Login
 ****************/
 func (app *App) Login(login_id string, password string) {
  //
}