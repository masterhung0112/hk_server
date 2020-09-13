/**
 * Contains a list of functions related to user
 * - User creation
 * - Soft-delete user
 */

package app

import (
	"fmt"
	"github.com/masterhung0112/go_server/mlog"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"net/http"
	"strings"
)

const (
	TOKEN_TYPE_PASSWORD_RECOVERY  = "password_recovery"
	TOKEN_TYPE_VERIFY_EMAIL       = "verify_email"
	TOKEN_TYPE_TEAM_INVITATION    = "team_invitation"
	TOKEN_TYPE_GUEST_INVITATION   = "guest_invitation"
	TOKEN_TYPE_CWS_ACCESS         = "cws_access_token"
	PASSWORD_RECOVER_EXPIRY_TIME  = 1000 * 60 * 60      // 1 hour
	INVITATION_EXPIRY_TIME        = 1000 * 60 * 60 * 48 // 48 hours
	IMAGE_PROFILE_PIXEL_DIMENSION = 128
)

func (a *App) GetUser(userId string) (*model.User, *model.AppError) {
	return a.Srv().Store.User().Get(userId)
}

func (a *App) GetUserByUsername(username string) (*model.User, *model.AppError) {
	result, err := a.Srv().Store.User().GetByUsername(username)
	if err != nil && err.Id == "store.sql_user.get_by_username.app_error" {
		err.StatusCode = http.StatusNotFound
		return nil, err
	}
	return result, nil
}

func (a *App) GetUserByEmail(email string) (*model.User, *model.AppError) {
	user, err := a.Srv().Store.User().GetByEmail(email)
	if err != nil {
		if err.Id == "store.sql_user.missing_account.const" {
			err.StatusCode = http.StatusNotFound
			return nil, err
		}
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}
	return user, nil
}

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

func (a *App) GetViewUsersRestrictions(userId string) (*model.ViewUsersRestrictions, *model.AppError) {
	if a.HasPermissionTo(userId, model.PERMISSION_VIEW_MEMBERS) {
		return nil, nil
	}

	teamIds, getTeamErr := a.Srv().Store.Team().GetUserTeamIds(userId, true)
	if getTeamErr != nil {
		return nil, getTeamErr
	}

	teamIdsWithPermission := []string{}
	for _, teamId := range teamIds {
		if a.HasPermissionToTeam(userId, teamId, model.PERMISSION_VIEW_MEMBERS) {
			teamIdsWithPermission = append(teamIdsWithPermission, teamId)
		}
	}

	userChannelMembers, err := a.Srv().Store.Channel().GetAllChannelMembersForUser(userId, true, true)
	if err != nil {
		return nil, err
	}

	channelIds := []string{}
	for channelId := range userChannelMembers {
		channelIds = append(channelIds, channelId)
	}

	return &model.ViewUsersRestrictions{Teams: teamIdsWithPermission, Channels: channelIds}, nil
}

func (a *App) GetUsersWithoutTeamPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersWithoutTeam(options)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersWithoutTeam(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfilesWithoutTeam(options)
}

func (a *App) sanitizeProfiles(users []*model.User, asAdmin bool) []*model.User {
	for _, u := range users {
		a.SanitizeProfile(u, asAdmin)
	}

	return users
}

func (a *App) GetSanitizeOptions(asAdmin bool) map[string]bool {
	options := a.Config().GetSanitizeOptions()
	if asAdmin {
		options["email"] = true
		options["fullname"] = true
		options["authservice"] = true
	}
	return options
}

func (a *App) SanitizeProfile(user *model.User, asAdmin bool) {
	options := a.GetSanitizeOptions(asAdmin)

	user.SanitizeProfile(options)
}

func (a *App) GetUsersNotInChannel(teamId string, channelId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfilesNotInChannel(teamId, channelId, groupConstrained, offset, limit, viewRestrictions)
}

func (a *App) GetUsersNotInChannelPage(teamId string, channelId string, groupConstrained bool, page int, perPage int, asAdmin bool, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersNotInChannel(teamId, channelId, groupConstrained, page*perPage, perPage, viewRestrictions)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfilesNotInTeam(teamId, groupConstrained, offset, limit, viewRestrictions)
}

func (a *App) GetUsersNotInTeamEtag(teamId string, restrictionsHash string) string {
	return fmt.Sprintf("%v.%v.%v.%v", a.Srv().Store.User().GetEtagForProfilesNotInTeam(teamId), a.Config().PrivacySettings.ShowFullName, a.Config().PrivacySettings.ShowEmailAddress, restrictionsHash)
}

func (a *App) GetUsersInTeamEtag(teamId string, restrictionsHash string) string {
	return fmt.Sprintf("%v.%v.%v.%v", a.Srv().Store.User().GetEtagForProfiles(teamId), a.Config().PrivacySettings.ShowFullName, a.Config().PrivacySettings.ShowEmailAddress, restrictionsHash)
}

func (a *App) GetUsersNotInTeamPage(teamId string, groupConstrained bool, page int, perPage int, asAdmin bool, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersNotInTeam(teamId, groupConstrained, page*perPage, perPage, viewRestrictions)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersInTeam(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfiles(options)
}

func (a *App) GetUsersInTeamPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInTeam(options)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersInChannelByStatus(channelId string, offset int, limit int) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfilesInChannelByStatus(channelId, offset, limit)
}

func (a *App) GetUsersInChannelPageByStatus(channelId string, page int, perPage int, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInChannelByStatus(channelId, page*perPage, perPage)
	if err != nil {
		return nil, err
	}
	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersInChannel(channelId string, offset int, limit int) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetProfilesInChannel(channelId, offset, limit)
}

func (a *App) GetUsersInChannelPage(channelId string, page int, perPage int, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInChannel(channelId, page*perPage, perPage)
	if err != nil {
		return nil, err
	}
	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsers(options)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsers(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	return a.Srv().Store.User().GetAllProfiles(options)
}

func (a *App) RestrictUsersGetByPermissions(userId string, options *model.UserGetOptions) (*model.UserGetOptions, *model.AppError) {
	restrictions, err := a.GetViewUsersRestrictions(userId)
	if err != nil {
		return nil, err
	}

	options.ViewRestrictions = restrictions
	return options, nil
}

func (a *App) UpdateUserRoles(userId string, newRoles string, sendWebSocketEvent bool) (*model.User, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		err.StatusCode = http.StatusBadRequest
		return nil, err
	}

	if err := a.CheckRolesExist(strings.Fields(newRoles)); err != nil {
		return nil, err
	}

	user.Roles = newRoles
	uchan := make(chan store.StoreResult, 1)
	go func() {
		userUpdate, err := a.Srv().Store.User().Update(user, true)
		uchan <- store.StoreResult{Data: userUpdate, Err: err}
		close(uchan)
	}()

	schan := make(chan store.StoreResult, 1)
	go func() {
		id, err := a.Srv().Store.Session().UpdateRoles(user.Id, newRoles)
		schan <- store.StoreResult{Data: id, NErr: err}
		close(schan)
	}()

	result := <-uchan
	if result.Err != nil {
		return nil, result.Err
	}
	ruser := result.Data.(*model.UserUpdate).New

	if result := <-schan; result.NErr != nil {
		// soft error since the user roles were still updated
		mlog.Error("Failed during updating user roles", mlog.Err(result.NErr))
	}

	//TODO: Open this
	// a.InvalidateCacheForUser(userId)
	// a.ClearSessionCacheForUser(user.Id)

	//TODO: Open this
	// if sendWebSocketEvent {
	// 	message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_ROLE_UPDATED, "", "", user.Id, nil)
	// 	message.Add("user_id", user.Id)
	// 	message.Add("roles", newRoles)
	// 	a.Publish(message)
	// }

	return ruser, nil
}
