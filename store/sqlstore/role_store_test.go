package sqlstore

import (
	"github.com/stretchr/testify/suite"
	"github.com/masterhung0112/go_server/model"
  "testing"
)

type RoleStoreTestSuite struct {
  suite.Suite
  StoreTestSuite
}

func (s *RoleStoreTestSuite) SetupTest() {
  // s.InitInitializeStore()
// 	// utils.TranslationsPreInit()

// 	// backend, err := NewFileBackend(&s.settings, true)
// 	// require.Nil(s.T(), err)
// 	// s.backend = backend
}

func TestRoleStoreTestSuite(t *testing.T) {
  StoreTestSuiteWithSqlSupplier(t, &RoleStoreTestSuite{})
}

// func TestRoleStore(t *testing.T) {
// 	StoreTestWithSqlSupplier(t, storetest.TestRoleStore)
// }

// func TestRoleStoreSave(t *testing.T) {
//   StoreTestWithSqlSupplier(t, testRoleStoreSave)
// }

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
}