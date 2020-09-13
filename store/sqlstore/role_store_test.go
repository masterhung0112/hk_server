package sqlstore

import (
	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RoleStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func TestRoleStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &RoleStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *RoleStoreTestSuite) TestRoleStoreSave() {
	// Save a new role.
	r1 := &model.Role{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Permissions: []string{
			"invite_user",
			"create_public_channel",
			"add_user_to_team",
		},
		SchemeManaged: false,
	}

	d1, err := s.Store().Role().Save(r1)
	if s.Nil(err) && s.NotNil(d1) {
		s.Len(d1.Id, 26)
		s.Equal(r1.Name, d1.Name)
		s.Equal(r1.DisplayName, d1.DisplayName)
		s.Equal(r1.Description, d1.Description)
		s.Equal(r1.Permissions, d1.Permissions)
		s.Equal(r1.SchemeManaged, d1.SchemeManaged)
	}

	// Change the role permissions and update.
	d1.Permissions = []string{
		"invite_user",
		"add_user_to_team",
		"delete_public_channel",
	}

	d2, err := s.Store().Role().Save(d1)
	if s.Nil(err) && s.NotNil(d2) {
		s.Len(d2.Id, 26)
		s.Equal(r1.Name, d2.Name)
		s.Equal(r1.DisplayName, d2.DisplayName)
		s.Equal(r1.Description, d2.Description)
		s.Equal(d1.Permissions, d2.Permissions)
		s.Equal(r1.SchemeManaged, d2.SchemeManaged)
	}

	// Try saving one with an invalid ID set.
	r3 := &model.Role{
		Id:          model.NewId(),
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Permissions: []string{
			"invite_user",
			"create_public_channel",
			"add_user_to_team",
		},
		SchemeManaged: false,
	}

	_, err = s.Store().Role().Save(r3)
	s.NotNil(err)

	// Try saving one with a duplicate "name" field.
	r4 := &model.Role{
		Name:        r1.Name,
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Permissions: []string{
			"invite_user",
			"create_public_channel",
			"add_user_to_team",
		},
		SchemeManaged: false,
	}

	_, err = s.Store().Role().Save(r4)
	s.NotNil(err)

	r5 := &model.Role{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Permissions: []string{
			"invalid_permission",
		},
		SchemeManaged: false,
	}

	_, err = s.Store().Role().Save(r5)
	if s.NotNil(err) {
		s.Contains(err.Error(), "invalid_permission")
	}
}

func (s *RoleStoreTestSuite) TestRoleStoreGetAll() {
	prev, err := s.Store().Role().GetAll()
	if s.Nil(err) {
		prevCount := len(prev)

		// Save a role to test with.
		r1 := &model.Role{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Permissions: []string{
				"invite_user",
				"create_public_channel",
				"add_user_to_team",
			},
			SchemeManaged: false,
		}

		_, err = s.Store().Role().Save(r1)
		s.Nil(err)

		r2 := &model.Role{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Permissions: []string{
				"invite_user",
				"create_public_channel",
				"add_user_to_team",
			},
			SchemeManaged: false,
		}
		_, err = s.Store().Role().Save(r2)
		s.Nil(err)

		data, err := s.Store().Role().GetAll()
		s.Nil(err)
		s.Len(data, prevCount+2)
	}
}

func (s *RoleStoreTestSuite) TestRoleStoreGetByName() {
	// Save a role to test with.
	r1 := &model.Role{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Permissions: []string{
			"invite_user",
			"create_public_channel",
			"add_user_to_team",
		},
		SchemeManaged: false,
	}

	d1, err := s.Store().Role().Save(r1)
	s.Require().Nil(err)
	s.Require().Len(d1.Id, 26)

	// Get a valid role
	d2, err := s.Store().Role().GetByName(d1.Name)
	s.Require().Nil(err)
	s.Require().Equal(d1.Id, d2.Id)
	s.Require().Equal(r1.Name, d2.Name)
	s.Require().Equal(r1.DisplayName, d2.DisplayName)
	s.Require().Equal(r1.Description, d2.Description)
	s.Require().Equal(r1.Permissions, d2.Permissions)
	s.Require().Equal(r1.SchemeManaged, d2.SchemeManaged)

	// Get an invalid role
	_, err = s.Store().Role().GetByName(model.NewId())
	s.Require().NotNil(err)
}
