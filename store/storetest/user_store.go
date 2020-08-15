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
  t.Run("Get", func(t *testing.T) { testUserStoreGet(t, ss) })
}

func testUserStoreSave(t *testing.T, ss store.Store) {
  // teamId := model.NewId()
  // maxUsersPerTeam := 50

  u1 := model.User{
		Email:    MakeEmail(),
		Username: model.NewId(),
  }

  // Save the user to DB
  _, err := ss.User().Save(&u1)
  require.Nil(t, err, "couldn't save user")

  // Delete the user after running the test
  defer func() { require.Nil(t, ss.User().PermanentDelete(u1.Id)) }()

  //TODO: Open
  // _, err = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, maxUsersPerTeam)
	// require.Nil(t, err)

	_, err = ss.User().Save(&u1)
  require.NotNil(t, err, "shouldn't be able to update user from save")

  u2 := model.User{
		Email:    u1.Email,
		Username: model.NewId(),
	}
	_, err = ss.User().Save(&u2)
  require.NotNil(t, err, "should be unique email")
  require.Equal(t, "store.sql_user.save.email_exists.app_error", err.Message)

  u2 = model.User{
    Email: MakeEmail(),
    Username: u1.Username,
  }
	_, err = ss.User().Save(&u2)
  require.NotNil(t, err, "should be unique username")
  require.Equal(t, "store.sql_user.save.username_exists.app_error", err.Message)

  // Username auto-generated if Username is empty
	u2 = model.User{
    Email: MakeEmail(),
    Username: "",
  }
	_, err = ss.User().Save(&u2)
  require.Nil(t, err, "Expect no error")

  for i := 0; i < 49; i++ {
    u := model.User{
			Email:    MakeEmail(),
			Username: model.NewId(),
		}
		_, err = ss.User().Save(&u)
    require.Nil(t, err, "couldn't save item")

    defer func() { require.Nil(t, ss.User().PermanentDelete(u.Id)) }()

    //TODO: test open
    // _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u.Id}, maxUsersPerTeam)
		// require.Nil(t, nErr)
  }

  u2 = model.User{
    Id: "",
	  Email: MakeEmail(),
    Username: model.NewId(),
  }
	_, err = ss.User().Save(&u2)
  require.Nil(t, err, "couldn't save item")

  defer func() { require.Nil(t, ss.User().PermanentDelete(u2.Id)) }()

  //TODO: test open
  // _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, maxUsersPerTeam)
	// require.NotNil(t, nErr, "should be the limit")
}

func testUserStoreGet(t *testing.T, ss store.Store) {
  // Make a user and save to DB
  u1 := &model.User{
		Email: MakeEmail(),
  }
  _, err := ss.User().Save(u1)
	require.Nil(t, err)
  defer func() { require.Nil(t, ss.User().PermanentDelete(u1.Id)) }()

  u2, _ := ss.User().Save(&model.User{
		Email:    MakeEmail(),
		Username: model.NewId(),
  })
  defer func() { require.Nil(t, ss.User().PermanentDelete(u2.Id)) }()

  //TODO: Open test
	// _, nErr := ss.Bot().Save(&model.Bot{
	// 	UserId:      u2.Id,
	// 	Username:    u2.Username,
	// 	Description: "bot description",
	// 	OwnerId:     u1.Id,
	// })
	// require.Nil(t, nErr)
	// u2.IsBot = true
	// u2.BotDescription = "bot description"
  // defer func() { require.Nil(t, ss.Bot().PermanentDelete(u2.Id)) }()

  // _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
  // require.Nil(t, nErr)

  t.Run("fetch empty id", func(t *testing.T) {
		_, err := ss.User().Get("")
		require.NotNil(t, err)
  })

  t.Run("fetch user 1", func(t *testing.T) {
		actual, err := ss.User().Get(u1.Id)
		require.Nil(t, err)
		require.Equal(t, u1, actual)
		// require.False(t, actual.IsBot)
	})

	t.Run("fetch user 2, also a bot", func(t *testing.T) {
		actual, err := ss.User().Get(u2.Id)
		require.Nil(t, err)
		require.Equal(t, u2, actual)
		// require.True(t, actual.IsBot)
		// require.Equal(t, "bot description", actual.BotDescription)
	})
}