/**
 * Contains a list of functions related to user
 * - User creation
 * - Soft-delete user
 */

package app

/****************
 * User Creation
 ****************/
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