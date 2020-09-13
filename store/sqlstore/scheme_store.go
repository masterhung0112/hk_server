package sqlstore

import (
	"fmt"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
)

type SqlSchemeStore struct {
	SqlStore
}

func newSqlSchemeStore(sqlStore SqlStore) store.SchemeStore {
	s := &SqlSchemeStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Scheme{}, "Schemes").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Name").SetMaxSize(model.SCHEME_NAME_MAX_LENGTH).SetUnique(true)
		table.ColMap("DisplayName").SetMaxSize(model.SCHEME_DISPLAY_NAME_MAX_LENGTH)
		table.ColMap("Description").SetMaxSize(model.SCHEME_DESCRIPTION_MAX_LENGTH)
		table.ColMap("Scope").SetMaxSize(32)
		table.ColMap("DefaultTeamAdminRole").SetMaxSize(64)
		table.ColMap("DefaultTeamUserRole").SetMaxSize(64)
		table.ColMap("DefaultTeamGuestRole").SetMaxSize(64)
		table.ColMap("DefaultChannelAdminRole").SetMaxSize(64)
		table.ColMap("DefaultChannelUserRole").SetMaxSize(64)
		table.ColMap("DefaultChannelGuestRole").SetMaxSize(64)
	}

	return s
}

func filterModerated(permissions []string) []string {
	filteredPermissions := []string{}
	for _, perm := range permissions {
		if _, ok := model.ChannelModeratedPermissionsMap[perm]; ok {
			filteredPermissions = append(filteredPermissions, perm)
		}
	}
	return filteredPermissions
}

func (s *SqlSchemeStore) createScheme(scheme *model.Scheme, transaction *gorp.Transaction) (*model.Scheme, error) {
	// Fetch the default system scheme roles to populate default permissions.
	defaultRoleNames := []string{model.TEAM_ADMIN_ROLE_ID, model.TEAM_USER_ROLE_ID, model.TEAM_GUEST_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID, model.CHANNEL_USER_ROLE_ID, model.CHANNEL_GUEST_ROLE_ID}
	defaultRoles := make(map[string]*model.Role)
	roles, appErr := s.SqlStore.Role().GetByNames(defaultRoleNames)
	if appErr != nil {
		return nil, appErr
	}

	for _, role := range roles {
		switch role.Name {
		case model.TEAM_ADMIN_ROLE_ID:
			defaultRoles[model.TEAM_ADMIN_ROLE_ID] = role
		case model.TEAM_USER_ROLE_ID:
			defaultRoles[model.TEAM_USER_ROLE_ID] = role
		case model.TEAM_GUEST_ROLE_ID:
			defaultRoles[model.TEAM_GUEST_ROLE_ID] = role
		case model.CHANNEL_ADMIN_ROLE_ID:
			defaultRoles[model.CHANNEL_ADMIN_ROLE_ID] = role
		case model.CHANNEL_USER_ROLE_ID:
			defaultRoles[model.CHANNEL_USER_ROLE_ID] = role
		case model.CHANNEL_GUEST_ROLE_ID:
			defaultRoles[model.CHANNEL_GUEST_ROLE_ID] = role
		}
	}

	if len(defaultRoles) != 6 {
		return nil, errors.New("createScheme: unable to retrieve default scheme roles")
	}

	// Create the appropriate default roles for the scheme.
	if scheme.Scope == model.SCHEME_SCOPE_TEAM {
		// Team Admin Role
		teamAdminRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team Admin Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_ADMIN_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err := s.SqlStore.Role().(*SqlRoleStore).createRole(teamAdminRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamAdminRole = savedRole.Name

		// Team User Role
		teamUserRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team User Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_USER_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(teamUserRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamUserRole = savedRole.Name

		// Team Guest Role
		teamGuestRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Team Guest Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.TEAM_GUEST_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(teamGuestRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultTeamGuestRole = savedRole.Name
	}

	if scheme.Scope == model.SCHEME_SCOPE_TEAM || scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
		// Channel Admin Role
		channelAdminRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel Admin Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_ADMIN_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelAdminRole.Permissions = []string{}
		}

		savedRole, err := s.SqlStore.Role().(*SqlRoleStore).createRole(channelAdminRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelAdminRole = savedRole.Name

		// Channel User Role
		channelUserRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel User Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_USER_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelUserRole.Permissions = filterModerated(channelUserRole.Permissions)
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(channelUserRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelUserRole = savedRole.Name

		// Channel Guest Role
		channelGuestRole := &model.Role{
			Name:          model.NewId(),
			DisplayName:   fmt.Sprintf("Channel Guest Role for Scheme %s", scheme.Name),
			Permissions:   defaultRoles[model.CHANNEL_GUEST_ROLE_ID].Permissions,
			SchemeManaged: true,
		}

		if scheme.Scope == model.SCHEME_SCOPE_CHANNEL {
			channelGuestRole.Permissions = filterModerated(channelGuestRole.Permissions)
		}

		savedRole, err = s.SqlStore.Role().(*SqlRoleStore).createRole(channelGuestRole, transaction)
		if err != nil {
			return nil, err
		}
		scheme.DefaultChannelGuestRole = savedRole.Name
	}

	scheme.Id = model.NewId()
	if len(scheme.Name) == 0 {
		scheme.Name = model.NewId()
	}
	scheme.CreateAt = model.GetMillis()
	scheme.UpdateAt = scheme.CreateAt

	// Validate the scheme
	if !scheme.IsValidForCreate() {
		return nil, store.NewErrInvalidInput("Scheme", "<any>", fmt.Sprintf("%v", scheme))
	}

	if err := transaction.Insert(scheme); err != nil {
		return nil, errors.Wrap(err, "failed to save Scheme")
	}

	return scheme, nil
}

func (s *SqlSchemeStore) Save(scheme *model.Scheme) (*model.Scheme, error) {
	if len(scheme.Id) == 0 {
		transaction, err := s.GetMaster().Begin()
		if err != nil {
			return nil, errors.Wrap(err, "begin_transaction")
		}
		defer finalizeTransaction(transaction)

		newScheme, err := s.createScheme(scheme, transaction)
		if err != nil {
			return nil, err
		}
		if err := transaction.Commit(); err != nil {
			return nil, errors.Wrap(err, "commit_transaction")
		}
		return newScheme, nil
	}

	if !scheme.IsValid() {
		return nil, store.NewErrInvalidInput("Scheme", "<any>", fmt.Sprintf("%v", scheme))
	}

	scheme.UpdateAt = model.GetMillis()

	rowsChanged, err := s.GetMaster().Update(scheme)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update Scheme")
	}
	if rowsChanged != 1 {
		return nil, errors.New("no record to update")
	}

	return scheme, nil
}

func (s *SqlSchemeStore) GetAllPage(scope string, offset int, limit int) ([]*model.Scheme, error) {
	var schemes []*model.Scheme

	scopeClause := ""
	if len(scope) > 0 {
		scopeClause = " AND Scope=:Scope "
	}

	if _, err := s.GetReplica().Select(&schemes, "SELECT * from Schemes WHERE DeleteAt = 0 "+scopeClause+" ORDER BY CreateAt DESC LIMIT :Limit OFFSET :Offset", map[string]interface{}{"Limit": limit, "Offset": offset, "Scope": scope}); err != nil {
		return nil, errors.Wrapf(err, "failed to get Schemes")
	}

	return schemes, nil
}
