package api

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/web"
	"net/http"
	"strconv"
	"strings"
)

func (api *API) InitUser() {
	api.BaseRoutes.Users.Handle("", api.ApiHandler(createUser)).Methods("POST")
	api.BaseRoutes.Users.Handle("", api.ApiSessionRequired(getUsers)).Methods("GET")

	api.BaseRoutes.User.Handle("", api.ApiSessionRequired(getUser)).Methods("GET")

	api.BaseRoutes.Users.Handle("/login", api.ApiHandler(login)).Methods("POST")
	api.BaseRoutes.Users.Handle("/logout", api.ApiHandler(logout)).Methods("POST")

}

func CreateUser(c *web.Context, w http.ResponseWriter, r *http.Request) {
	// Convert Json to User model
	user := model.UserFromJson(r.Body)
	if user == nil {
		c.SetInvalidParam("user")
		return
	}

	user.SanitizeInput(c.IsSystemAdmin())

	redirect := r.URL.Query().Get("r")

	ruser, err := c.App.CreateUserFromSignup(user, redirect)

	if err != nil {
		return
	}

	// Successfully created new user
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ruser.ToJson()))
}

func createUser(c *Context, w http.ResponseWriter, r *http.Request) {
	user := model.UserFromJson(r.Body)
	if user == nil {
		c.SetInvalidParam("user")
		return
	}

	redirect := r.URL.Query().Get("r")

	var ruser *model.User
	ruser, err := c.App.CreateUserFromSignup(user, redirect)

	if err != nil {
		c.Err = err
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(ruser.ToJson()))
}

func getUsers(c *Context, w http.ResponseWriter, r *http.Request) {
	inTeamId := r.URL.Query().Get("in_team")
	notInTeamId := r.URL.Query().Get("not_in_team")
	inChannelId := r.URL.Query().Get("in_channel")
	inGroupId := r.URL.Query().Get("in_group")
	notInChannelId := r.URL.Query().Get("not_in_channel")
	groupConstrained := r.URL.Query().Get("group_constrained")
	withoutTeam := r.URL.Query().Get("without_team")
	inactive := r.URL.Query().Get("inactive")
	active := r.URL.Query().Get("active")
	role := r.URL.Query().Get("role")
	sort := r.URL.Query().Get("sort")
	rolesString := r.URL.Query().Get("roles")
	channelRolesString := r.URL.Query().Get("channel_roles")
	teamRolesString := r.URL.Query().Get("team_roles")

	if len(notInChannelId) > 0 && len(inTeamId) == 0 {
		c.SetInvalidUrlParam("team_id")
		return
	}

	if sort != "" && sort != "last_activity_at" && sort != "create_at" && sort != "status" {
		c.SetInvalidUrlParam("sort")
		return
	}

	// Currently only supports sorting on a team
	// or sort="status" on inChannelId
	if (sort == "last_activity_at" || sort == "create_at") && (inTeamId == "" || notInTeamId != "" || inChannelId != "" || notInChannelId != "" || withoutTeam != "" || inGroupId != "") {
		c.SetInvalidUrlParam("sort")
		return
	}
	if sort == "status" && inChannelId == "" {
		c.SetInvalidUrlParam("sort")
		return
	}

	withoutTeamBool, _ := strconv.ParseBool(withoutTeam)
	groupConstrainedBool, _ := strconv.ParseBool(groupConstrained)
	inactiveBool, _ := strconv.ParseBool(inactive)
	activeBool, _ := strconv.ParseBool(active)

	if inactiveBool && activeBool {
		c.SetInvalidUrlParam("inactive")
	}

	roles := []string{}
	var rolesValid bool
	if rolesString != "" {
		roles, rolesValid = model.CleanRoleNames(strings.Split(rolesString, ","))
		if !rolesValid {
			c.SetInvalidParam("roles")
			return
		}
	}
	channelRoles := []string{}
	if channelRolesString != "" && len(inChannelId) != 0 {
		channelRoles, rolesValid = model.CleanRoleNames(strings.Split(channelRolesString, ","))
		if !rolesValid {
			c.SetInvalidParam("channelRoles")
			return
		}
	}
	teamRoles := []string{}
	if teamRolesString != "" && len(inTeamId) != 0 {
		teamRoles, rolesValid = model.CleanRoleNames(strings.Split(teamRolesString, ","))
		if !rolesValid {
			c.SetInvalidParam("teamRoles")
			return
		}
	}

	restrictions, err := c.App.GetViewUsersRestrictions(c.App.Session().UserId)
	if err != nil {
		c.Err = err
		return
	}

	userGetOptions := &model.UserGetOptions{
		InTeamId:         inTeamId,
		InChannelId:      inChannelId,
		NotInTeamId:      notInTeamId,
		NotInChannelId:   notInChannelId,
		InGroupId:        inGroupId,
		GroupConstrained: groupConstrainedBool,
		WithoutTeam:      withoutTeamBool,
		Inactive:         inactiveBool,
		Active:           activeBool,
		Role:             role,
		Roles:            roles,
		ChannelRoles:     channelRoles,
		TeamRoles:        teamRoles,
		Sort:             sort,
		Page:             c.Params.Page,
		PerPage:          c.Params.PerPage,
		ViewRestrictions: restrictions,
	}

	var profiles []*model.User
	etag := ""

	if withoutTeamBool, _ := strconv.ParseBool(withoutTeam); withoutTeamBool {
		// Use a special permission for now
		if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_LIST_USERS_WITHOUT_TEAM) {
			c.SetPermissionError(model.PERMISSION_LIST_USERS_WITHOUT_TEAM)
			return
		}

		profiles, err = c.App.GetUsersWithoutTeamPage(userGetOptions, c.IsSystemAdmin())
	} else if len(notInChannelId) > 0 {
		if !c.App.SessionHasPermissionToChannel(*c.App.Session(), notInChannelId, model.PERMISSION_READ_CHANNEL) {
			c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
			return
		}

		profiles, err = c.App.GetUsersNotInChannelPage(inTeamId, notInChannelId, groupConstrainedBool, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin(), restrictions)
	} else if len(notInTeamId) > 0 {
		if !c.App.SessionHasPermissionToTeam(*c.App.Session(), notInTeamId, model.PERMISSION_VIEW_TEAM) {
			c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
			return
		}

		etag = c.App.GetUsersNotInTeamEtag(inTeamId, restrictions.Hash())
		if c.HandleEtag(etag, "Get Users Not in Team", w, r) {
			return
		}

		profiles, err = c.App.GetUsersNotInTeamPage(notInTeamId, groupConstrainedBool, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin(), restrictions)
	} else if len(inTeamId) > 0 {
		if !c.App.SessionHasPermissionToTeam(*c.App.Session(), inTeamId, model.PERMISSION_VIEW_TEAM) {
			c.SetPermissionError(model.PERMISSION_VIEW_TEAM)
			return
		}

		if sort == "last_activity_at" {
			//TODO: Open this
			// profiles, err = c.App.GetRecentlyActiveUsersForTeamPage(inTeamId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin(), restrictions)
		} else if sort == "create_at" {
			//TODO: Open this
			// profiles, err = c.App.GetNewUsersForTeamPage(inTeamId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin(), restrictions)
		} else {
			etag = c.App.GetUsersInTeamEtag(inTeamId, restrictions.Hash())
			if c.HandleEtag(etag, "Get Users in Team", w, r) {
				return
			}
			profiles, err = c.App.GetUsersInTeamPage(userGetOptions, c.IsSystemAdmin())
		}
	} else if len(inChannelId) > 0 {
		if !c.App.SessionHasPermissionToChannel(*c.App.Session(), inChannelId, model.PERMISSION_READ_CHANNEL) {
			c.SetPermissionError(model.PERMISSION_READ_CHANNEL)
			return
		}
		if sort == "status" {
			profiles, err = c.App.GetUsersInChannelPageByStatus(userGetOptions, c.IsSystemAdmin())
		} else {
			profiles, err = c.App.GetUsersInChannelPage(userGetOptions, c.IsSystemAdmin())
		}
	} else if len(inGroupId) > 0 {
		//TODO: Open this
		// if c.App.Srv().License() == nil || !*c.App.Srv().License().Features.LDAPGroups {
		// 	c.Err = model.NewAppError("Api1.getUsersInGroup", "api.ldap_groups.license_error", nil, "", http.StatusNotImplemented)
		// 	return
		// }

		if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS) {
			c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS)
			return
		}

		//TODO: Open this
		// profiles, _, err = c.App.GetGroupMemberUsersPage(inGroupId, c.Params.Page, c.Params.PerPage)
		// if err != nil {
		// 	c.Err = err
		// 	return
		// }
	} else {
		userGetOptions, err = c.App.RestrictUsersGetByPermissions(c.App.Session().UserId, userGetOptions)
		if err != nil {
			c.Err = err
			return
		}
		profiles, err = c.App.GetUsersPage(userGetOptions, c.IsSystemAdmin())
	}

	if err != nil {
		c.Err = err
		return
	}

	if len(etag) > 0 {
		w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	}

	//TODO: Open this
	// c.App.UpdateLastActivityAtIfNeeded(*c.App.Session())
	w.Write([]byte(model.UserListToJson(profiles)))
}

func logout(c *Context, w http.ResponseWriter, r *http.Request) {
	Logout(c, w, r)
}

func Logout(c *Context, w http.ResponseWriter, r *http.Request) {
	//TODO: Open
	// auditRec := c.MakeAuditRecord("Logout", audit.Fail)
	// defer c.LogAuditRec(auditRec)
	// c.LogAudit("")

	c.RemoveSessionCookie(w, r)
	if c.App.Session().Id != "" {
		if err := c.App.RevokeSessionById(c.App.Session().Id); err != nil {
			c.Err = err
			return
		}
	}

	//TODO: Open
	// auditRec.Success()
	ReturnStatusOK(w)
}

func login(c *Context, w http.ResponseWriter, r *http.Request) {
	// Mask all sensitive errors, with the exception of the following
	defer func() {
		if c.Err == nil {
			return
		}

		unmaskedErrors := []string{
			"mfa.validate_token.authenticate.app_error",
			"api.user.check_user_mfa.bad_code.app_error",
			"api.user.login.blank_pwd.app_error",
			"api.user.login.bot_login_forbidden.app_error",
			"api.user.login.client_side_cert.certificate.app_error",
			"api.user.login.inactive.app_error",
			"api.user.login.not_verified.app_error",
			"api.user.check_user_login_attempts.too_many.app_error",
			"app.team.join_user_to_team.max_accounts.app_error",
			"store.sql_user.save.max_accounts.app_error",
		}

		maskError := true

		for _, unmaskedError := range unmaskedErrors {
			if c.Err.Id == unmaskedError {
				maskError = false
			}
		}

		if !maskError {
			return
		}

		config := c.App.Config()
		enableUsername := *config.EmailSettings.EnableSignInWithUsername
		enableEmail := *config.EmailSettings.EnableSignInWithEmail
		//TODO: Open
		// samlEnabled := *config.SamlSettings.Enable
		// gitlabEnabled := *config.GetSSOService("gitlab").Enable
		// googleEnabled := *config.GetSSOService("google").Enable
		// office365Enabled := *config.Office365Settings.Enable

		// if samlEnabled || gitlabEnabled || googleEnabled || office365Enabled {
		// 	c.Err = model.NewAppError("login", "api.user.login.invalid_credentials_sso", nil, "", http.StatusUnauthorized)
		// 	return
		// }

		if enableUsername && !enableEmail {
			c.Err = model.NewAppError("login", "api.user.login.invalid_credentials_username", nil, "", http.StatusUnauthorized)
			return
		}

		if !enableUsername && enableEmail {
			c.Err = model.NewAppError("login", "api.user.login.invalid_credentials_email", nil, "", http.StatusUnauthorized)
			return
		}

		c.Err = model.NewAppError("login", "api.user.login.invalid_credentials_email_username", nil, "", http.StatusUnauthorized)
	}()

	props := model.MapFromJson(r.Body)
	id := props["id"]
	loginId := props["login_id"]
	password := props["password"]
	mfaToken := props["token"]
	deviceId := props["device_id"]
	ldapOnly := props["ldap_only"] == "true"

	//TODO: Open
	// if *c.App.Config().ExperimentalSettings.ClientSideCertEnable {
	// 	if license := c.App.Srv().License(); license == nil || !*license.Features.SAML {
	// 		c.Err = model.NewAppError("ClientSideCertNotAllowed", "api.user.login.client_side_cert.license.app_error", nil, "", http.StatusBadRequest)
	// 		return
	// 	}
	// 	certPem, certSubject, certEmail := c.App.CheckForClientSideCert(r)
	// 	mlog.Debug("Client Cert", mlog.String("cert_subject", certSubject), mlog.String("cert_email", certEmail))

	// 	if len(certPem) == 0 || len(certEmail) == 0 {
	// 		c.Err = model.NewAppError("ClientSideCertMissing", "api.user.login.client_side_cert.certificate.app_error", nil, "", http.StatusBadRequest)
	// 		return
	// 	}

	// 	if *c.App.Config().ExperimentalSettings.ClientSideCertCheck == model.CLIENT_SIDE_CERT_CHECK_PRIMARY_AUTH {
	// 		loginId = certEmail
	// 		password = "certificate"
	// 	}
	// }

	// auditRec := c.MakeAuditRecord("login", audit.Fail)
	// defer c.LogAuditRec(auditRec)
	// auditRec.AddMeta("login_id", loginId)
	// auditRec.AddMeta("device_id", deviceId)

	// c.LogAuditWithUserId(id, "attempt - login_id="+loginId)

	user, err := c.App.AuthenticateUserForLogin(id, loginId, password, mfaToken, "", ldapOnly)
	if err != nil {
		// c.LogAuditWithUserId(id, "failure - login_id="+loginId)
		c.Err = err
		return
	}
	// auditRec.AddMeta("user", user)

	if user.IsGuest() {
		if c.App.Srv().License() == nil {
			c.Err = model.NewAppError("login", "api.user.login.guest_accounts.license.error", nil, "", http.StatusUnauthorized)
			return
		}
		if !*c.App.Config().GuestAccountsSettings.Enable {
			c.Err = model.NewAppError("login", "api.user.login.guest_accounts.disabled.error", nil, "", http.StatusUnauthorized)
			return
		}
	}

	// c.LogAuditWithUserId(user.Id, "authenticated")

	err = c.App.DoLogin(w, r, user, deviceId, false, false, false)
	if err != nil {
		c.Err = err
		return
	}

	// c.LogAuditWithUserId(user.Id, "success")

	if r.Header.Get(model.HEADER_REQUESTED_WITH) == model.HEADER_REQUESTED_WITH_XML {
		c.App.AttachSessionCookies(w, r)
	}

	//TODO: Open
	// userTermsOfService, err := c.App.GetUserTermsOfService(user.Id)
	// if err != nil && err.StatusCode != http.StatusNotFound {
	// 	c.Err = err
	// 	return
	// }

	// if userTermsOfService != nil {
	// 	user.TermsOfServiceId = userTermsOfService.TermsOfServiceId
	// 	user.TermsOfServiceCreateAt = userTermsOfService.CreateAt
	// }

	user.Sanitize(map[string]bool{})

	// auditRec.Success()
	w.Write([]byte(user.ToJson()))
}

func getUser(c *Context, w http.ResponseWriter, r *http.Request) {
	c.RequireUserId()
	if c.Err != nil {
		return
	}

	canSee, err := c.App.UserCanSeeOtherUser(c.App.Session().UserId, c.Params.UserId)
	if err != nil {
		c.SetPermissionError(model.PERMISSION_VIEW_MEMBERS)
		return
	}

	if !canSee {
		c.SetPermissionError(model.PERMISSION_VIEW_MEMBERS)
		return
	}

	user, err := c.App.GetUser(c.Params.UserId)
	if err != nil {
		c.Err = err
		return
	}

	if c.IsSystemAdmin() || c.App.Session().UserId == user.Id {
		//TODO: Open
		// userTermsOfService, err := c.App.GetUserTermsOfService(user.Id)
		// if err != nil && err.StatusCode != http.StatusNotFound {
		// 	c.Err = err
		// 	return
		// }

		// if userTermsOfService != nil {
		// 	user.TermsOfServiceId = userTermsOfService.TermsOfServiceId
		// 	user.TermsOfServiceCreateAt = userTermsOfService.CreateAt
		// }
	}

	//TODO: Open
	// etag := user.Etag(*c.App.Config().PrivacySettings.ShowFullName, *c.App.Config().PrivacySettings.ShowEmailAddress)

	// if c.HandleEtag(etag, "Get User", w, r) {
	// 	return
	// }

	if c.App.Session().UserId == user.Id {
		user.Sanitize(map[string]bool{})
	} else {
		c.App.SanitizeProfile(user, c.IsSystemAdmin())
	}
	//TODO: Open
	// c.App.UpdateLastActivityAtIfNeeded(*c.App.Session())
	// w.Header().Set(model.HEADER_ETAG_SERVER, etag)
	w.Write([]byte(user.ToJson()))
}
