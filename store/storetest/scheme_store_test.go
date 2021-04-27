package storetest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/store"
)

type SchemeStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func TestSchemeStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &SchemeStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *SchemeStoreTestSuite) SetupSuite() {
	createDefaultRoles(s.Store())
}

func createDefaultRoles(s store.Store) {
	s.Role().Save(&model.Role{
		Name:        model.TEAM_ADMIN_ROLE_ID,
		DisplayName: model.TEAM_ADMIN_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_DELETE_OTHERS_POSTS.Id,
		},
	})

	s.Role().Save(&model.Role{
		Name:        model.TEAM_USER_ROLE_ID,
		DisplayName: model.TEAM_USER_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_VIEW_TEAM.Id,
			model.PERMISSION_ADD_USER_TO_TEAM.Id,
		},
	})

	s.Role().Save(&model.Role{
		Name:        model.TEAM_GUEST_ROLE_ID,
		DisplayName: model.TEAM_GUEST_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_VIEW_TEAM.Id,
		},
	})

	s.Role().Save(&model.Role{
		Name:        model.CHANNEL_ADMIN_ROLE_ID,
		DisplayName: model.CHANNEL_ADMIN_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_MANAGE_PUBLIC_CHANNEL_MEMBERS.Id,
			model.PERMISSION_MANAGE_PRIVATE_CHANNEL_MEMBERS.Id,
		},
	})

	s.Role().Save(&model.Role{
		Name:        model.CHANNEL_USER_ROLE_ID,
		DisplayName: model.CHANNEL_USER_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_READ_CHANNEL.Id,
			model.PERMISSION_CREATE_POST.Id,
		},
	})

	s.Role().Save(&model.Role{
		Name:        model.CHANNEL_GUEST_ROLE_ID,
		DisplayName: model.CHANNEL_GUEST_ROLE_ID,
		Permissions: []string{
			model.PERMISSION_READ_CHANNEL.Id,
			model.PERMISSION_CREATE_POST.Id,
		},
	})
}

func (s *SchemeStoreTestSuite) TestSchemeStoreSave() {
	// Save a new scheme.
	s1 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	// Check all fields saved correctly.
	d1, err := s.Store().Scheme().Save(s1)
	s.Assert().Nil(err)
	s.Assert().Len(d1.Id, 26)
	s.Assert().Equal(s1.DisplayName, d1.DisplayName)
	s.Assert().Equal(s1.Name, d1.Name)
	s.Assert().Equal(s1.Description, d1.Description)
	s.Assert().NotZero(d1.CreateAt)
	s.Assert().NotZero(d1.UpdateAt)
	s.Assert().Zero(d1.DeleteAt)
	s.Assert().Equal(s1.Scope, d1.Scope)
	s.Assert().Len(d1.DefaultTeamAdminRole, 26)
	s.Assert().Len(d1.DefaultTeamUserRole, 26)
	s.Assert().Len(d1.DefaultTeamGuestRole, 26)
	s.Assert().Len(d1.DefaultChannelAdminRole, 26)
	s.Assert().Len(d1.DefaultChannelUserRole, 26)
	s.Assert().Len(d1.DefaultChannelGuestRole, 26)

	// Check the default roles were created correctly.
	role1, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamAdminRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role1.Permissions, []string{"delete_others_posts"})
	s.Assert().True(role1.SchemeManaged)

	role2, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamUserRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role2.Permissions, []string{"view_team", "add_user_to_team"})
	s.Assert().True(role2.SchemeManaged)

	role3, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelAdminRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role3.Permissions, []string{"manage_public_channel_members", "manage_private_channel_members"})
	s.Assert().True(role3.SchemeManaged)

	role4, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelUserRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role4.Permissions, []string{"read_channel", "create_post"})
	s.Assert().True(role4.SchemeManaged)

	role5, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamGuestRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role5.Permissions, []string{"view_team"})
	s.Assert().True(role5.SchemeManaged)

	role6, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelGuestRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role6.Permissions, []string{"read_channel", "create_post"})
	s.Assert().True(role6.SchemeManaged)

	// Change the scheme description and update.
	d1.Description = model.NewId()

	d2, err := s.Store().Scheme().Save(d1)
	s.Assert().Nil(err)
	s.Assert().Equal(d1.Id, d2.Id)
	s.Assert().Equal(s1.DisplayName, d2.DisplayName)
	s.Assert().Equal(s1.Name, d2.Name)
	s.Assert().Equal(d1.Description, d2.Description)
	s.Assert().NotZero(d2.CreateAt)
	s.Assert().NotZero(d2.UpdateAt)
	s.Assert().Zero(d2.DeleteAt)
	s.Assert().Equal(s1.Scope, d2.Scope)
	s.Assert().Equal(d1.DefaultTeamAdminRole, d2.DefaultTeamAdminRole)
	s.Assert().Equal(d1.DefaultTeamUserRole, d2.DefaultTeamUserRole)
	s.Assert().Equal(d1.DefaultTeamGuestRole, d2.DefaultTeamGuestRole)
	s.Assert().Equal(d1.DefaultChannelAdminRole, d2.DefaultChannelAdminRole)
	s.Assert().Equal(d1.DefaultChannelUserRole, d2.DefaultChannelUserRole)
	s.Assert().Equal(d1.DefaultChannelGuestRole, d2.DefaultChannelGuestRole)

	// Try saving one with an invalid ID set.
	s3 := &model.Scheme{
		Id:          model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	_, err = s.Store().Scheme().Save(s3)
	s.Assert().NotNil(err)
}

func (s *SchemeStoreTestSuite) TestSchemeStoreGet() {
	// Save a scheme to test with.
	s1 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	d1, err := s.Store().Scheme().Save(s1)
	s.Assert().Nil(err)
	s.Assert().Len(d1.Id, 26)

	// Get a valid scheme
	d2, err := s.Store().Scheme().Get(d1.Id)
	s.Assert().Nil(err)
	s.Assert().Equal(d1.Id, d2.Id)
	s.Assert().Equal(s1.DisplayName, d2.DisplayName)
	s.Assert().Equal(s1.Name, d2.Name)
	s.Assert().Equal(d1.Description, d2.Description)
	s.Assert().NotZero(d2.CreateAt)
	s.Assert().NotZero(d2.UpdateAt)
	s.Assert().Zero(d2.DeleteAt)
	s.Assert().Equal(s1.Scope, d2.Scope)
	s.Assert().Equal(d1.DefaultTeamAdminRole, d2.DefaultTeamAdminRole)
	s.Assert().Equal(d1.DefaultTeamUserRole, d2.DefaultTeamUserRole)
	s.Assert().Equal(d1.DefaultTeamGuestRole, d2.DefaultTeamGuestRole)
	s.Assert().Equal(d1.DefaultChannelAdminRole, d2.DefaultChannelAdminRole)
	s.Assert().Equal(d1.DefaultChannelUserRole, d2.DefaultChannelUserRole)
	s.Assert().Equal(d1.DefaultChannelGuestRole, d2.DefaultChannelGuestRole)

	// Get an invalid scheme
	_, err = s.Store().Scheme().Get(model.NewId())
	s.Assert().NotNil(err)
}

func (s *SchemeStoreTestSuite) TestSchemeStoreGetByName() {
	// Save a scheme to test with.
	s1 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	d1, err := s.Store().Scheme().Save(s1)
	s.Assert().Nil(err)
	s.Assert().Len(d1.Id, 26)

	// Get a valid scheme
	d2, err := s.Store().Scheme().GetByName(d1.Name)
	s.Assert().Nil(err)
	s.Assert().Equal(d1.Id, d2.Id)
	s.Assert().Equal(s1.DisplayName, d2.DisplayName)
	s.Assert().Equal(s1.Name, d2.Name)
	s.Assert().Equal(d1.Description, d2.Description)
	s.Assert().NotZero(d2.CreateAt)
	s.Assert().NotZero(d2.UpdateAt)
	s.Assert().Zero(d2.DeleteAt)
	s.Assert().Equal(s1.Scope, d2.Scope)
	s.Assert().Equal(d1.DefaultTeamAdminRole, d2.DefaultTeamAdminRole)
	s.Assert().Equal(d1.DefaultTeamUserRole, d2.DefaultTeamUserRole)
	s.Assert().Equal(d1.DefaultTeamGuestRole, d2.DefaultTeamGuestRole)
	s.Assert().Equal(d1.DefaultChannelAdminRole, d2.DefaultChannelAdminRole)
	s.Assert().Equal(d1.DefaultChannelUserRole, d2.DefaultChannelUserRole)
	s.Assert().Equal(d1.DefaultChannelGuestRole, d2.DefaultChannelGuestRole)

	// Get an invalid scheme
	_, err = s.Store().Scheme().GetByName(model.NewId())
	s.Assert().NotNil(err)
}

func (s *SchemeStoreTestSuite) TestSchemeStoreGetAllPage() {
	// Save a scheme to test with.
	schemes := []*model.Scheme{
		{
			DisplayName: model.NewId(),
			Name:        model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		},
		{
			DisplayName: model.NewId(),
			Name:        model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		},
		{
			DisplayName: model.NewId(),
			Name:        model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		},
		{
			DisplayName: model.NewId(),
			Name:        model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		},
	}

	for _, scheme := range schemes {
		_, err := s.Store().Scheme().Save(scheme)
		s.Require().Nil(err)
	}

	s1, err := s.Store().Scheme().GetAllPage("", 0, 2)
	s.Assert().Nil(err)
	s.Assert().Len(s1, 2)

	s2, err := s.Store().Scheme().GetAllPage("", 2, 2)
	s.Assert().Nil(err)
	s.Assert().Len(s2, 2)
	s.Assert().NotEqual(s1[0].DisplayName, s2[0].DisplayName)
	s.Assert().NotEqual(s1[0].DisplayName, s2[1].DisplayName)
	s.Assert().NotEqual(s1[1].DisplayName, s2[0].DisplayName)
	s.Assert().NotEqual(s1[1].DisplayName, s2[1].DisplayName)
	s.Assert().NotEqual(s1[0].Name, s2[0].Name)
	s.Assert().NotEqual(s1[0].Name, s2[1].Name)
	s.Assert().NotEqual(s1[1].Name, s2[0].Name)
	s.Assert().NotEqual(s1[1].Name, s2[1].Name)

	s3, err := s.Store().Scheme().GetAllPage("team", 0, 1000)
	s.Assert().Nil(err)
	s.Assert().NotZero(len(s3))
	for _, data := range s3 {
		s.Assert().Equal("team", data.Scope)
	}

	s4, err := s.Store().Scheme().GetAllPage("channel", 0, 1000)
	s.Assert().Nil(err)
	s.Assert().NotZero(len(s4))
	for _, data := range s4 {
		s.Assert().Equal("channel", data.Scope)
	}
}

func (s *SchemeStoreTestSuite) TestSchemeStoreDelete() {
	// Save a new scheme.
	s1 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	// Check all fields saved correctly.
	d1, err := s.Store().Scheme().Save(s1)
	s.Assert().Nil(err)
	s.Assert().Len(d1.Id, 26)
	s.Assert().Equal(s1.DisplayName, d1.DisplayName)
	s.Assert().Equal(s1.Name, d1.Name)
	s.Assert().Equal(s1.Description, d1.Description)
	s.Assert().NotZero(d1.CreateAt)
	s.Assert().NotZero(d1.UpdateAt)
	s.Assert().Zero(d1.DeleteAt)
	s.Assert().Equal(s1.Scope, d1.Scope)
	s.Assert().Len(d1.DefaultTeamAdminRole, 26)
	s.Assert().Len(d1.DefaultTeamUserRole, 26)
	s.Assert().Len(d1.DefaultTeamGuestRole, 26)
	s.Assert().Len(d1.DefaultChannelAdminRole, 26)
	s.Assert().Len(d1.DefaultChannelUserRole, 26)
	s.Assert().Len(d1.DefaultChannelGuestRole, 26)

	// Check the default roles were created correctly.
	role1, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamAdminRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role1.Permissions, []string{"delete_others_posts"})
	s.Assert().True(role1.SchemeManaged)

	role2, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamUserRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role2.Permissions, []string{"view_team", "add_user_to_team"})
	s.Assert().True(role2.SchemeManaged)

	role3, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelAdminRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role3.Permissions, []string{"manage_public_channel_members", "manage_private_channel_members"})
	s.Assert().True(role3.SchemeManaged)

	role4, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelUserRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role4.Permissions, []string{"read_channel", "create_post"})
	s.Assert().True(role4.SchemeManaged)

	role5, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamGuestRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role5.Permissions, []string{"view_team"})
	s.Assert().True(role5.SchemeManaged)

	role6, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelGuestRole)
	s.Assert().Nil(err)
	s.Assert().Equal(role6.Permissions, []string{"read_channel", "create_post"})
	s.Assert().True(role6.SchemeManaged)

	// Delete the scheme.
	d2, err := s.Store().Scheme().Delete(d1.Id)
	s.Require().Nil(err)
	s.Assert().NotZero(d2.DeleteAt)

	// Check that the roles are deleted too.
	role7, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamAdminRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role7.DeleteAt)

	role8, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamUserRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role8.DeleteAt)

	role9, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelAdminRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role9.DeleteAt)

	role10, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelUserRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role10.DeleteAt)

	role11, err := s.Store().Role().GetByName(context.Background(), d1.DefaultTeamGuestRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role11.DeleteAt)

	role12, err := s.Store().Role().GetByName(context.Background(), d1.DefaultChannelGuestRole)
	s.Assert().Nil(err)
	s.Assert().NotZero(role12.DeleteAt)

	// Try deleting a scheme that does not exist.
	_, err = s.Store().Scheme().Delete(model.NewId())
	s.Assert().NotNil(err)

	// Try deleting a team scheme that's in use.
	s4 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}
	d4, err := s.Store().Scheme().Save(s4)
	s.Assert().Nil(err)

	t4 := &model.Team{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Email:       MakeEmail(),
		Type:        model.TEAM_OPEN,
		SchemeId:    &d4.Id,
	}
	t4, err = s.Store().Team().Save(t4)
	s.Require().Nil(err)

	_, err = s.Store().Scheme().Delete(d4.Id)
	s.Assert().Nil(err)

	t5, err := s.Store().Team().Get(t4.Id)
	s.Require().Nil(err)
	s.Assert().Equal("", *t5.SchemeId)

	// Try deleting a channel scheme that's in use.
	s5 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}
	d5, err := s.Store().Scheme().Save(s5)
	s.Assert().Nil(err)

	c5 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &d5.Id,
	}
	c5, nErr := s.Store().Channel().Save(c5, -1)
	s.Assert().Nil(nErr)

	_, err = s.Store().Scheme().Delete(d5.Id)
	s.Assert().Nil(err)

	c6, nErr := s.Store().Channel().Get(c5.Id, true)
	s.Assert().Nil(nErr)
	s.Assert().Equal("", *c6.SchemeId)
}

func (s *SchemeStoreTestSuite) TestSchemeStorePermanentDeleteAll() {
	s1 := &model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_TEAM,
	}

	s2 := &model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}

	s1, err := s.Store().Scheme().Save(s1)
	s.Require().Nil(err)
	s2, err = s.Store().Scheme().Save(s2)
	s.Require().Nil(err)

	err = s.Store().Scheme().PermanentDeleteAll()
	s.Assert().Nil(err)

	_, err = s.Store().Scheme().Get(s1.Id)
	s.Assert().NotNil(err)

	_, err = s.Store().Scheme().Get(s2.Id)
	s.Assert().NotNil(err)

	schemes, err := s.Store().Scheme().GetAllPage("", 0, 100000)
	s.Assert().Nil(err)
	s.Assert().Empty(schemes)
}

func (s *SchemeStoreTestSuite) TestSchemeStoreCountByScope() {
	testCounts := func(expectedTeamCount, expectedChannelCount int) {
		actualCount, err := s.Store().Scheme().CountByScope(model.SCHEME_SCOPE_TEAM)
		s.Require().Nil(err)
		s.Require().Equal(int64(expectedTeamCount), actualCount)

		actualCount, err = s.Store().Scheme().CountByScope(model.SCHEME_SCOPE_CHANNEL)
		s.Require().Nil(err)
		s.Require().Equal(int64(expectedChannelCount), actualCount)
	}

	createScheme := func(scope string) {
		_, err := s.Store().Scheme().Save(&model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       scope,
		})
		s.Require().Nil(err)
	}

	err := s.Store().Scheme().PermanentDeleteAll()
	s.Require().Nil(err)

	createScheme(model.SCHEME_SCOPE_CHANNEL)
	createScheme(model.SCHEME_SCOPE_TEAM)
	testCounts(1, 1)
	createScheme(model.SCHEME_SCOPE_TEAM)
	testCounts(2, 1)
	createScheme(model.SCHEME_SCOPE_CHANNEL)
	testCounts(2, 2)
}

func (s *SchemeStoreTestSuite) TestCountWithoutPermission() {
	perm := model.PERMISSION_CREATE_POST.Id

	createScheme := func(scope string) *model.Scheme {
		scheme, err := s.Store().Scheme().Save(&model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       scope,
		})
		s.Require().Nil(err)
		return scheme
	}

	getRoles := func(scheme *model.Scheme) (channelUser, channelGuest *model.Role) {
		var err error
		channelUser, err = s.Store().Role().GetByName(context.Background(), scheme.DefaultChannelUserRole)
		s.Require().Nil(err)
		s.Require().NotNil(channelUser)
		channelGuest, err = s.Store().Role().GetByName(context.Background(), scheme.DefaultChannelGuestRole)
		s.Require().Nil(err)
		s.Require().NotNil(channelGuest)
		return
	}

	teamScheme1 := createScheme(model.SCHEME_SCOPE_TEAM)
	defer s.Store().Scheme().Delete(teamScheme1.Id)
	teamScheme2 := createScheme(model.SCHEME_SCOPE_TEAM)
	defer s.Store().Scheme().Delete(teamScheme2.Id)
	channelScheme1 := createScheme(model.SCHEME_SCOPE_CHANNEL)
	defer s.Store().Scheme().Delete(channelScheme1.Id)
	channelScheme2 := createScheme(model.SCHEME_SCOPE_CHANNEL)
	defer s.Store().Scheme().Delete(channelScheme2.Id)

	ts1User, ts1Guest := getRoles(teamScheme1)
	ts2User, ts2Guest := getRoles(teamScheme2)
	cs1User, cs1Guest := getRoles(channelScheme1)
	cs2User, cs2Guest := getRoles(channelScheme2)

	allRoles := []*model.Role{
		ts1User,
		ts1Guest,
		ts2User,
		ts2Guest,
		cs1User,
		cs1Guest,
		cs2User,
		cs2Guest,
	}

	teamUserCount, err := s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_TEAM, perm, model.RoleScopeChannel, model.RoleTypeUser)
	s.Require().Nil(err)
	s.Require().Equal(int64(0), teamUserCount)

	teamGuestCount, err := s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_TEAM, perm, model.RoleScopeChannel, model.RoleTypeGuest)
	s.Require().Nil(err)
	s.Require().Equal(int64(0), teamGuestCount)

	var tests = []struct {
		removePermissionFromRole             *model.Role
		expectTeamSchemeChannelUserCount     int
		expectTeamSchemeChannelGuestCount    int
		expectChannelSchemeChannelUserCount  int
		expectChannelSchemeChannelGuestCount int
	}{
		{ts1User, 1, 0, 0, 0},
		{ts1Guest, 1, 1, 0, 0},
		{ts2User, 2, 1, 0, 0},
		{ts2Guest, 2, 2, 0, 0},
		{cs1User, 2, 2, 1, 0},
		{cs1Guest, 2, 2, 1, 1},
		{cs2User, 2, 2, 2, 1},
		{cs2Guest, 2, 2, 2, 2},
	}

	removePermission := func(targetRole *model.Role) {
		roleMatched := false
		for _, role := range allRoles {
			if targetRole == role {
				roleMatched = true
				role.Permissions = []string{}
				_, err = s.Store().Role().Save(role)
				s.Require().Nil(err)
			}
		}
		s.Require().True(roleMatched)
	}

	for _, test := range tests {
		removePermission(test.removePermissionFromRole)

		count, err := s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_TEAM, perm, model.RoleScopeChannel, model.RoleTypeUser)
		s.Require().Nil(err)
		s.Require().Equal(int64(test.expectTeamSchemeChannelUserCount), count)

		count, err = s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_TEAM, perm, model.RoleScopeChannel, model.RoleTypeGuest)
		s.Require().Nil(err)
		s.Require().Equal(int64(test.expectTeamSchemeChannelGuestCount), count)

		count, err = s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, perm, model.RoleScopeChannel, model.RoleTypeUser)
		s.Require().Nil(err)
		s.Require().Equal(int64(test.expectChannelSchemeChannelUserCount), count)

		count, err = s.Store().Scheme().CountWithoutPermission(model.SCHEME_SCOPE_CHANNEL, perm, model.RoleScopeChannel, model.RoleTypeGuest)
		s.Require().Nil(err)
		s.Require().Equal(int64(test.expectChannelSchemeChannelGuestCount), count)
	}
}
