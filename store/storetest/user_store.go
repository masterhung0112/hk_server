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
	t.Run("Count", func(t *testing.T) { testCount(t, ss) })
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
		Email:    MakeEmail(),
		Username: u1.Username,
	}
	_, err = ss.User().Save(&u2)
	require.NotNil(t, err, "should be unique username")
	require.Equal(t, "store.sql_user.save.username_exists.app_error", err.Message)

	// Username auto-generated if Username is empty
	u2 = model.User{
		Email:    MakeEmail(),
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
		Id:       "",
		Email:    MakeEmail(),
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

func testCount(t *testing.T, ss store.Store) {
	// teamId := model.NewId()
	// channelId := model.NewId()
	regularUser := &model.User{}
	regularUser.Email = MakeEmail()
	regularUser.Roles = model.SYSTEM_USER_ROLE_ID

	_, err := ss.User().Save(regularUser)
	require.Nil(t, err)
	defer func() { require.Nil(t, ss.User().PermanentDelete(regularUser.Id)) }()
	// _, nErr := ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: regularUser.Id, SchemeAdmin: false, SchemeUser: true}, -1)
	// require.Nil(t, nErr)
	// _, err = ss.Channel().SaveMember(&model.ChannelMember{UserId: regularUser.Id, ChannelId: channelId, SchemeAdmin: false, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	// require.Nil(t, err)

	guestUser := &model.User{}
	guestUser.Email = MakeEmail()
	guestUser.Roles = model.SYSTEM_GUEST_ROLE_ID
	_, err = ss.User().Save(guestUser)
	require.Nil(t, err)
	defer func() { require.Nil(t, ss.User().PermanentDelete(guestUser.Id)) }()
	// _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: guestUser.Id, SchemeAdmin: false, SchemeUser: false, SchemeGuest: true}, -1)
	// require.Nil(t, nErr)
	// _, err = ss.Channel().SaveMember(&model.ChannelMember{UserId: guestUser.Id, ChannelId: channelId, SchemeAdmin: false, SchemeUser: false, SchemeGuest: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	// require.Nil(t, err)

	// teamAdmin := &model.User{}
	// teamAdmin.Email = MakeEmail()
	// teamAdmin.Roles = model.SYSTEM_USER_ROLE_ID
	// _, err = ss.User().Save(teamAdmin)
	// require.Nil(t, err)
	// defer func() { require.Nil(t, ss.User().PermanentDelete(teamAdmin.Id)) }()
	// _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: teamAdmin.Id, SchemeAdmin: true, SchemeUser: true}, -1)
	// require.Nil(t, nErr)
	// _, err = ss.Channel().SaveMember(&model.ChannelMember{UserId: teamAdmin.Id, ChannelId: channelId, SchemeAdmin: true, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	// require.Nil(t, err)

	sysAdmin := &model.User{}
	sysAdmin.Email = MakeEmail()
	sysAdmin.Roles = model.SYSTEM_ADMIN_ROLE_ID + " " + model.SYSTEM_USER_ROLE_ID
	_, err = ss.User().Save(sysAdmin)
	require.Nil(t, err)
	defer func() { require.Nil(t, ss.User().PermanentDelete(sysAdmin.Id)) }()
	// _, nErr = ss.Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: sysAdmin.Id, SchemeAdmin: false, SchemeUser: true}, -1)
	// require.Nil(t, nErr)
	// _, err = ss.Channel().SaveMember(&model.ChannelMember{UserId: sysAdmin.Id, ChannelId: channelId, SchemeAdmin: true, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	// require.Nil(t, err)

	// Deleted
	deletedUser := &model.User{}
	deletedUser.Email = MakeEmail()
	deletedUser.DeleteAt = model.GetMillis()
	_, err = ss.User().Save(deletedUser)
	require.Nil(t, err)
	defer func() { require.Nil(t, ss.User().PermanentDelete(deletedUser.Id)) }()

	// Bot
	// botUser, err := ss.User().Save(&model.User{
	// 	Email: MakeEmail(),
	// })
	// require.Nil(t, err)
	// defer func() { require.Nil(t, ss.User().PermanentDelete(botUser.Id)) }()
	// _, nErr = ss.Bot().Save(&model.Bot{
	// 	UserId:   botUser.Id,
	// 	Username: botUser.Username,
	// 	OwnerId:  regularUser.Id,
	// })
	// require.Nil(t, nErr)
	// botUser.IsBot = true
	// defer func() { require.Nil(t, ss.Bot().PermanentDelete(botUser.Id)) }()

	testCases := []struct {
		Description string
		Options     model.UserCountOptions
		Expected    int64
	}{
		{
			"No bot accounts no deleted accounts and no team id",
			model.UserCountOptions{
				IncludeBotAccounts: false,
				IncludeDeleted:     false,
				TeamId:             "",
			},
			4,
		},
		// {
		// 	"Include bot accounts no deleted accounts and no team id",
		// 	model.UserCountOptions{
		// 		IncludeBotAccounts: true,
		// 		IncludeDeleted:     false,
		// 		TeamId:             "",
		// 	},
		// 	5,
		// },
		//TODO: Add more case
	}

	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			count, err := ss.User().Count(testCase.Options)
			require.Nil(t, err)
			require.Equal(t, testCase.Expected, count)
		})
	}
}
