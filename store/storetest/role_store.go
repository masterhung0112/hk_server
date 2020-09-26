package storetest

import (
	"github.com/masterhung0112/hk_server/store"
	"github.com/stretchr/testify/suite"
	"testing"
)

type RoleStoreTestSuite struct {
	suite.Suite
}

func (s *RoleStoreTestSuite) SetupTest() {
}

func TestRoleStore(t *testing.T, ss store.Store, s SqlSupplier) {

}

// func testRoleStoreSave(t *testing.T) {
//   // Save a new role.
// 	r1 := &model.Role{
// 		Name:        model.NewId(),
// 		DisplayName: model.NewId(),
// 		Description: model.NewId(),
// 		Permissions: []string{
// 			"invite_user",
// 			"create_public_channel",
// 			"add_user_to_team",
// 		},
// 		SchemeManaged: false,
// 	}

// 	d1, err := ss.Role().Save(r1)
// 	assert.Nil(t, err)
// 	assert.Len(t, d1.Id, 26)
// 	assert.Equal(t, r1.Name, d1.Name)
// 	assert.Equal(t, r1.DisplayName, d1.DisplayName)
// 	assert.Equal(t, r1.Description, d1.Description)
// 	assert.Equal(t, r1.Permissions, d1.Permissions)
// 	assert.Equal(t, r1.SchemeManaged, d1.SchemeManaged)
// }
