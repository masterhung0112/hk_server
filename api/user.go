package api

import (
	"strings"
	"strconv"
  "net/http"
  "github.com/masterhung0112/go_server/model"
  "github.com/masterhung0112/go_server/web"
)

func (api *API) InitUser() {
  api.BaseRoutes.Users.Handle("", api.ApiHandler(createUser)).Methods("POST")
  api.BaseRoutes.Users.Handle("", api.ApiHandler(getUsers)).Methods("GET")
}


func CreateUser(c *web.Context, w http.ResponseWriter, r *http.Request) {
  // Convert Json to User model
  user := model.UserFromJson(r.Body)

  ruser, err := c.App.CreateUserFromSignup(user)

  if err != nil {
    return
  }

  // Successfully created new user
  w.WriteHeader(http.StatusCreated)
  w.Write([]byte(ruser.ToJson()))
}

func createUser(c *Context, w http.ResponseWriter, r *http.Request) {
  user := model.UserFromJson(r.Body)

  var ruser *model.User
  ruser, err := c.App.CreateUserFromSignup(user)

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
			profiles, err = c.App.GetUsersInChannelPageByStatus(inChannelId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
		} else {
			profiles, err = c.App.GetUsersInChannelPage(inChannelId, c.Params.Page, c.Params.PerPage, c.IsSystemAdmin())
		}
	} else if len(inGroupId) > 0 {
		if c.App.Srv().License() == nil || !*c.App.Srv().License().Features.LDAPGroups {
			c.Err = model.NewAppError("Api4.getUsersInGroup", "api.ldap_groups.license_error", nil, "", http.StatusNotImplemented)
			return
		}

		if !c.App.SessionHasPermissionTo(*c.App.Session(), model.PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS) {
			c.SetPermissionError(model.PERMISSION_SYSCONSOLE_READ_USERMANAGEMENT_GROUPS)
			return
		}

		profiles, _, err = c.App.GetGroupMemberUsersPage(inGroupId, c.Params.Page, c.Params.PerPage)
		if err != nil {
			c.Err = err
			return
		}
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
	c.App.UpdateLastActivityAtIfNeeded(*c.App.Session())
	w.Write([]byte(model.UserListToJson(profiles)))
}