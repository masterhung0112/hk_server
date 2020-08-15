package storetest

import (
	"github.com/masterhung0112/go_server/model"
	"testing"

	"github.com/masterhung0112/go_server/store"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T, ss store.Store, s SqlSupplier) {
	users, err := ss.User().GetAll()
	require.Nil(t, err, "failed cleaning up test users")

	for _, u := range users {
		err := ss.User().PermanentDelete(u.Id)
		require.Nil(t, err, "failed cleaning up test user %s", u.Username)
  }

  t.Run("Save", func(t *testing.T) { testUserStoreSave(t, ss) })
}

func testUserStoreSave(t *testing.T, ss store.Store) {
  // teamId := model.NewId()
  // maxUsersPerTeam := 50

  u1 := model.User{
		Email:    MakeEmail(),
		Username: model.NewId(),
  }

  _, err := ss.User().Save(&u1)
	require.Nil(t, err, "couldn't save user")
}
