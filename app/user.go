/**
 * Contains a list of functions related to user
 * - User creation
 * - Soft-delete user
 */

package app

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"image"
	"image/png"
	"io"
	"mime/multipart"
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
	user, err := a.Srv().Store.User().Get(userId)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetUser", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetUser", "app.user.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	return user, nil
}

func (a *App) GetUserByUsername(username string) (*model.User, *model.AppError) {
	result, err := a.Srv().Store.User().GetByUsername(username)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetUserByUsername", "app.user.get_by_username.app_error", nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetUserByUsername", "app.user.get_by_username.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}
	return result, nil
}

func (a *App) GetUserByEmail(email string) (*model.User, *model.AppError) {
	user, err := a.Srv().Store.User().GetByEmail(email)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("GetUserByEmail", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("GetUserByEmail", MISSING_ACCOUNT_ERROR, nil, err.Error(), http.StatusInternalServerError)
		}
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
		return nil, model.NewAppError("createUserOrGuest", "app.user.get_total_users_count.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	if count <= 0 {
		user.Roles = model.SYSTEM_ADMIN_ROLE_ID + " " + model.SYSTEM_USER_ROLE_ID
	}

	// if _, ok := utils.GetSupportedLocales()[user.Locale]; !ok {
	// 	user.Locale = *a.Config().LocalizationSettings.DefaultClientLocale
	// }

	ruser, appErr := app.createUser(user)
	if appErr != nil {
		return nil, appErr
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

	ruser, nErr := a.Srv().Store.User().Save(user)
	if nErr != nil {
		mlog.Error("Couldn't save the user", mlog.Err(nErr))
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(nErr, &appErr):
			return nil, appErr
		case errors.As(nErr, &invErr):
			switch invErr.Field {
			case "email":
				return nil, model.NewAppError("createUser", "app.user.save.email_exists.app_error", nil, invErr.Error(), http.StatusBadRequest)
			case "username":
				return nil, model.NewAppError("createUser", "app.user.save.username_exists.app_error", nil, invErr.Error(), http.StatusBadRequest)
			default:
				return nil, model.NewAppError("createUser", "app.user.save.existing.app_error", nil, invErr.Error(), http.StatusBadRequest)
			}
		default:
			return nil, model.NewAppError("createUser", "app.user.save.app_error", nil, nErr.Error(), http.StatusInternalServerError)
		}
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
		return nil, model.NewAppError("GetViewUsersRestrictions", "app.channel.get_channels.get.app_error", nil, err.Error(), http.StatusInternalServerError)
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
	users, err := a.Srv().Store.User().GetProfilesWithoutTeam(options)
	if err != nil {
		return nil, model.NewAppError("GetUsersWithoutTeam", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
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
	users, err := a.Srv().Store.User().GetProfilesNotInChannel(teamId, channelId, groupConstrained, offset, limit, viewRestrictions)
	if err != nil {
		return nil, model.NewAppError("GetUsersNotInChannel", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) GetUsersNotInChannelPage(teamId string, channelId string, groupConstrained bool, page int, perPage int, asAdmin bool, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersNotInChannel(teamId, channelId, groupConstrained, page*perPage, perPage, viewRestrictions)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersNotInTeam(teamId string, groupConstrained bool, offset int, limit int, viewRestrictions *model.ViewUsersRestrictions) ([]*model.User, *model.AppError) {
	users, err := a.Srv().Store.User().GetProfilesNotInTeam(teamId, groupConstrained, offset, limit, viewRestrictions)
	if err != nil {
		return nil, model.NewAppError("GetUsersNotInTeam", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
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
	users, err := a.Srv().Store.User().GetProfiles(options)
	if err != nil {
		return nil, model.NewAppError("GetUsersInTeam", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) GetUsersInTeamPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInTeam(options)
	if err != nil {
		return nil, err
	}

	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersInChannelByStatus(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	users, err := a.Srv().Store.User().GetProfilesInChannelByStatus(options)
	if err != nil {
		return nil, model.NewAppError("GetUsersInChannelByStatus", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) GetUsersInChannelPageByStatus(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInChannelByStatus(options)
	if err != nil {
		return nil, err
	}
	return a.sanitizeProfiles(users, asAdmin), nil
}

func (a *App) GetUsersInChannel(options *model.UserGetOptions) ([]*model.User, *model.AppError) {
	users, err := a.Srv().Store.User().GetProfilesInChannel(options)
	if err != nil {
		return nil, model.NewAppError("GetUsersInChannel", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) GetUsersInChannelPage(options *model.UserGetOptions, asAdmin bool) ([]*model.User, *model.AppError) {
	users, err := a.GetUsersInChannel(options)
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
	users, err := a.Srv().Store.User().GetAllProfiles(options)
	if err != nil {
		return nil, model.NewAppError("GetUsers", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
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
		uchan <- store.StoreResult{Data: userUpdate, NErr: err}
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

// FilterNonGroupChannelMembers returns the subset of the given user IDs of the users who are not members of groups
// associated to the channel excluding bots
func (a *App) FilterNonGroupChannelMembers(userIds []string, channel *model.Channel) ([]string, error) {
	channelGroupUsers, err := a.GetChannelGroupUsers(channel.Id)
	if err != nil {
		return nil, err
	}
	return a.filterNonGroupUsers(userIds, channelGroupUsers)
}

// filterNonGroupUsers is a helper function that takes a list of user ids and a list of users
// and returns the list of normal users present in userIds but not in groupUsers.
func (a *App) filterNonGroupUsers(userIds []string, groupUsers []*model.User) ([]string, error) {
	nonMemberIds := []string{}
	users, err := a.Srv().Store.User().GetProfileByIds(userIds, nil, false)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		userIsMember := user.IsBot

		for _, pu := range groupUsers {
			if pu.Id == user.Id {
				userIsMember = true
				break
			}
		}
		if !userIsMember {
			nonMemberIds = append(nonMemberIds, user.Id)
		}
	}

	return nonMemberIds, nil
}

// GetChannelGroupUsers returns the users who are associated to the channel via GroupChannels and GroupMembers.
func (a *App) GetChannelGroupUsers(channelID string) ([]*model.User, *model.AppError) {
	users, err := a.Srv().Store.User().GetChannelGroupUsers(channelID)
	if err != nil {
		return nil, model.NewAppError("GetChannelGroupUsers", "app.user.get_profiles.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return users, nil
}

func (a *App) UserCanSeeOtherUser(userId string, otherUserId string) (bool, *model.AppError) {
	if userId == otherUserId {
		return true, nil
	}

	restrictions, err := a.GetViewUsersRestrictions(userId)
	if err != nil {
		return false, err
	}

	if restrictions == nil {
		return true, nil
	}

	if len(restrictions.Teams) > 0 {
		result, err := a.Srv().Store.Team().UserBelongsToTeams(otherUserId, restrictions.Teams)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}

	if len(restrictions.Channels) > 0 {
		result, err := a.userBelongsToChannels(otherUserId, restrictions.Channels)
		if err != nil {
			return false, err
		}
		if result {
			return true, nil
		}
	}

	return false, nil
}

func (a *App) userBelongsToChannels(userId string, channelIds []string) (bool, *model.AppError) {
	return a.Srv().Store.Channel().UserBelongsToChannels(userId, channelIds)
}

// CheckEmailDomain checks that an email domain matches a list of space-delimited domains as a string.
func CheckEmailDomain(email string, domains string) bool {
	if len(domains) == 0 {
		return true
	}

	domainArray := strings.Fields(strings.TrimSpace(strings.ToLower(strings.Replace(strings.Replace(domains, "@", " ", -1), ",", " ", -1))))

	for _, d := range domainArray {
		if strings.HasSuffix(strings.ToLower(email), "@"+d) {
			return true
		}
	}

	return false
}

// CheckUserDomain checks that a user's email domain matches a list of space-delimited domains as a string.
func CheckUserDomain(user *model.User, domains string) bool {
	return CheckEmailDomain(user.Email, domains)
}

func (a *App) UpdateUser(user *model.User, sendNotifications bool) (*model.User, *model.AppError) {
	prev, err := a.Srv().Store.User().Get(user.Id)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return nil, model.NewAppError("UpdateUser", MISSING_ACCOUNT_ERROR, nil, nfErr.Error(), http.StatusNotFound)
		default:
			return nil, model.NewAppError("UpdateUser", "app.user.get.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	var newEmail string
	if user.Email != prev.Email {
		if !CheckUserDomain(user, *a.Config().TeamSettings.RestrictCreationToDomains) {
			if !prev.IsGuest() && !prev.IsLDAPUser() && !prev.IsSAMLUser() && user.Email != prev.Email {
				return nil, model.NewAppError("UpdateUser", "api.user.update_user.accepted_domain.app_error", nil, "", http.StatusBadRequest)
			}
		}

		if !CheckUserDomain(user, *a.Config().GuestAccountsSettings.RestrictCreationToDomains) {
			if prev.IsGuest() && !prev.IsLDAPUser() && !prev.IsSAMLUser() && user.Email != prev.Email {
				return nil, model.NewAppError("UpdateUser", "api.user.update_user.accepted_guest_domain.app_error", nil, "", http.StatusBadRequest)
			}
		}

		if _, appErr := a.GetUserByEmail(user.Email); appErr == nil {
			return nil, model.NewAppError("UpdateUser", "store.sql_user.update.email_taken.app_error", nil, "user_id="+user.Id, http.StatusBadRequest)
		}

		// Don't set new eMail on user account if email verification is required, this will be done as a post-verification action
		// to avoid users being able to set non-controlled eMails as their account email
		if *a.Config().EmailSettings.RequireEmailVerification && prev.Email != user.Email {
			newEmail = user.Email

			//  When a bot is created, prev.Email will be an autogenerated faked email,
			//  which will not match a CLI email input during bot to user conversions.
			//  To update a bot users email, do not set the email to the faked email
			//  stored in prev.Email.  Allow using the email defined in the CLI
			if !user.IsBot {
				user.Email = prev.Email
			}
		}
	}

	userUpdate, err := a.Srv().Store.User().Update(user, false)
	if err != nil {
		var appErr *model.AppError
		var invErr *store.ErrInvalidInput
		switch {
		case errors.As(err, &appErr):
			return nil, appErr
		case errors.As(err, &invErr):
			return nil, model.NewAppError("UpdateUser", "app.user.update.find.app_error", nil, invErr.Error(), http.StatusBadRequest)
		default:
			return nil, model.NewAppError("UpdateUser", "app.user.update.finding.app_error", nil, err.Error(), http.StatusInternalServerError)
		}
	}

	if sendNotifications {
		if newEmail != "" {

		}
		//TODO: Open
		// if userUpdate.New.Email != userUpdate.Old.Email || newEmail != "" {
		// 	if *a.Config().EmailSettings.RequireEmailVerification {
		// 		a.Srv().Go(func() {
		// 			if err := a.SendEmailVerification(userUpdate.New, newEmail, ""); err != nil {
		// 				mlog.Error("Failed to send email verification", mlog.Err(err))
		// 			}
		// 		})
		// 	} else {
		// 		a.Srv().Go(func() {
		// 			if err := a.Srv().EmailService.sendEmailChangeEmail(userUpdate.Old.Email, userUpdate.New.Email, userUpdate.New.Locale, a.GetSiteURL()); err != nil {
		// 				mlog.Error("Failed to send email change email", mlog.Err(err))
		// 			}
		// 		})
		// 	}
		// }

		// if userUpdate.New.Username != userUpdate.Old.Username {
		// 	a.Srv().Go(func() {
		// 		if err := a.Srv().EmailService.sendChangeUsernameEmail(userUpdate.Old.Username, userUpdate.New.Username, userUpdate.New.Email, userUpdate.New.Locale, a.GetSiteURL()); err != nil {
		// 			mlog.Error("Failed to send change username email", mlog.Err(err))
		// 		}
		// 	})
		// }
	}

	a.InvalidateCacheForUser(user.Id)

	return userUpdate.New, nil
}

func (a *App) UpdateUserNotifyProps(userId string, props map[string]string) (*model.User, *model.AppError) {
	user, err := a.GetUser(userId)
	if err != nil {
		return nil, err
	}

	user.NotifyProps = props

	ruser, err := a.UpdateUser(user, true)
	if err != nil {
		return nil, err
	}

	return ruser, nil
}

func (a *App) UpdatePassword(user *model.User, newPassword string) *model.AppError {
	if err := a.IsPasswordValid(newPassword); err != nil {
		return err
	}

	hashedPassword := model.HashPassword(newPassword)

	if err := a.Srv().Store.User().UpdatePassword(user.Id, hashedPassword); err != nil {
		return model.NewAppError("UpdatePassword", "api.user.update_password.failed.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	a.InvalidateCacheForUser(user.Id)

	return nil
}

func (a *App) SetProfileImageFromMultiPartFile(userId string, file multipart.File) *model.AppError {
	// Decode image config first to check dimensions before loading the whole thing into memory later on
	config, _, err := image.DecodeConfig(file)
	if err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.decode_config.app_error", nil, err.Error(), http.StatusBadRequest)
	}
	// This casting is done to prevent overflow on 32 bit systems (not needed
	// in 64 bits systems because images can't have more than 32 bits height or
	// width)
	if int64(config.Width)*int64(config.Height) > model.MaxImageSize {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.too_large.app_error", nil, "", http.StatusBadRequest)
	}

	file.Seek(0, 0)

	return a.SetProfileImageFromFile(userId, file)
}

func (a *App) SetProfileImageFromFile(userId string, file io.Reader) *model.AppError {

	buf, err := a.AdjustImage(file)
	if err != nil {
		return err
	}
	path := "users/" + userId + "/profile.png"

	if _, err := a.WriteFile(buf, path); err != nil {
		return model.NewAppError("SetProfileImage", "api.user.upload_profile_user.upload_profile.app_error", nil, "", http.StatusInternalServerError)
	}

	if err := a.Srv().Store.User().UpdateLastPictureUpdate(userId); err != nil {
		mlog.Error("Error with updating last picture update", mlog.Err(err))
	}
	a.invalidateUserCacheAndPublish(userId)

	return nil
}

func (a *App) AdjustImage(file io.Reader) (*bytes.Buffer, *model.AppError) {
	// Decode image into Image object
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, model.NewAppError("SetProfileImage", "api.user.upload_profile_user.decode.app_error", nil, err.Error(), http.StatusBadRequest)
	}

	orientation, _ := getImageOrientation(file)
	img = makeImageUpright(img, orientation)

	// Scale profile image
	profileWidthAndHeight := 128
	img = imaging.Fill(img, profileWidthAndHeight, profileWidthAndHeight, imaging.Center, imaging.Lanczos)

	buf := new(bytes.Buffer)
	err = png.Encode(buf, img)
	if err != nil {
		return nil, model.NewAppError("SetProfileImage", "api.user.upload_profile_user.encode.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	return buf, nil
}

// invalidateUserCacheAndPublish Invalidates cache for a user and publishes user updated event
func (a *App) invalidateUserCacheAndPublish(userId string) {
	a.InvalidateCacheForUser(userId)

	user, userErr := a.GetUser(userId)
	if userErr != nil {
		mlog.Error("Error in getting users profile", mlog.String("user_id", userId), mlog.Err(userErr))
		return
	}

	options := a.Config().GetSanitizeOptions()
	user.SanitizeProfile(options)

	//TODO: Open
	// message := model.NewWebSocketEvent(model.WEBSOCKET_EVENT_USER_UPDATED, "", "", "", nil)
	// message.Add("user", user)
	// a.Publish(message)
}
