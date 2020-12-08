package sqlstore

import (
	"time"
	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/masterhung0112/hk_server/store/storetest"
)

// type UserStoreTestSuite struct {
// 	suite.Suite
// 	StoreTestSuite
// }

func sanitized(user *model.User) *model.User {
	clonedUser := user.DeepCopy()
	clonedUser.Sanitize(map[string]bool{})

	return clonedUser
}

// func TestUserStoreTestSuite(t *testing.T) {
// 	StoreTestSuiteWithSqlSupplier(t, &UserStoreTestSuite{})
// }

// func TestUserStore(t *testing.T) {
// 	StoreTestWithSqlSupplier(t, storetest.TestUserStore)
// }

type UserStoreTS struct {
	suite.Suite
	StoreTestSuite
}

func TestUserStoreTS(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &UserStoreTS{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *UserStoreTS) TestCount() {
	teamId := model.NewId()
	channelId := model.NewId()
	regularUser := &model.User{}
	regularUser.Email = MakeEmail()
	regularUser.Roles = model.SYSTEM_USER_ROLE_ID
	_, err := s.Store().User().Save(regularUser)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(regularUser.Id)) }()
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: regularUser.Id, SchemeAdmin: false, SchemeUser: true}, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Channel().SaveMember(&model.ChannelMember{UserId: regularUser.Id, ChannelId: channelId, SchemeAdmin: false, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	s.Require().Nil(nErr)

	guestUser := &model.User{}
	guestUser.Email = MakeEmail()
	guestUser.Roles = model.SYSTEM_GUEST_ROLE_ID
	_, err = s.Store().User().Save(guestUser)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(guestUser.Id)) }()
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: guestUser.Id, SchemeAdmin: false, SchemeUser: false, SchemeGuest: true}, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Channel().SaveMember(&model.ChannelMember{UserId: guestUser.Id, ChannelId: channelId, SchemeAdmin: false, SchemeUser: false, SchemeGuest: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	s.Require().Nil(nErr)

	teamAdmin := &model.User{}
	teamAdmin.Email = MakeEmail()
	teamAdmin.Roles = model.SYSTEM_USER_ROLE_ID
	_, err = s.Store().User().Save(teamAdmin)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(teamAdmin.Id)) }()
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: teamAdmin.Id, SchemeAdmin: true, SchemeUser: true}, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Channel().SaveMember(&model.ChannelMember{UserId: teamAdmin.Id, ChannelId: channelId, SchemeAdmin: true, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	s.Require().Nil(nErr)

	sysAdmin := &model.User{}
	sysAdmin.Email = MakeEmail()
	sysAdmin.Roles = model.SYSTEM_ADMIN_ROLE_ID + " " + model.SYSTEM_USER_ROLE_ID
	_, err = s.Store().User().Save(sysAdmin)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(sysAdmin.Id)) }()
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: sysAdmin.Id, SchemeAdmin: false, SchemeUser: true}, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Channel().SaveMember(&model.ChannelMember{UserId: sysAdmin.Id, ChannelId: channelId, SchemeAdmin: true, SchemeUser: true, NotifyProps: model.GetDefaultChannelNotifyProps()})
	s.Require().Nil(nErr)

	// Deleted
	deletedUser := &model.User{}
	deletedUser.Email = MakeEmail()
	deletedUser.DeleteAt = model.GetMillis()
	_, err = s.Store().User().Save(deletedUser)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(deletedUser.Id)) }()

	// Bot
	botUser, err := s.Store().User().Save(&model.User{
		Email: MakeEmail(),
	})
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(botUser.Id)) }()
	_, nErr = s.Store().Bot().Save(&model.Bot{
		UserId:   botUser.Id,
		Username: botUser.Username,
		OwnerId:  regularUser.Id,
	})
	s.Require().Nil(nErr)
	botUser.IsBot = true
	defer func() { s.Require().Nil(s.Store().Bot().PermanentDelete(botUser.Id)) }()

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
		{
			"Include bot accounts no deleted accounts and no team id",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     false,
				TeamId:             "",
			},
			5,
		},
		{
			"Include delete accounts no bots and no team id",
			model.UserCountOptions{
				IncludeBotAccounts: false,
				IncludeDeleted:     true,
				TeamId:             "",
			},
			5,
		},
		{
			"Include bot accounts and deleted accounts and no team id",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     true,
				TeamId:             "",
			},
			6,
		},
		{
			"Include bot accounts, deleted accounts, exclude regular users with no team id",
			model.UserCountOptions{
				IncludeBotAccounts:  true,
				IncludeDeleted:      true,
				ExcludeRegularUsers: true,
				TeamId:              "",
			},
			1,
		},
		{
			"Include bot accounts and deleted accounts with existing team id",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     true,
				TeamId:             teamId,
			},
			4,
		},
		{
			"Include bot accounts and deleted accounts with fake team id",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     true,
				TeamId:             model.NewId(),
			},
			0,
		},
		{
			"Include bot accounts and deleted accounts with existing team id and view restrictions allowing team",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     true,
				TeamId:             teamId,
				ViewRestrictions:   &model.ViewUsersRestrictions{Teams: []string{teamId}},
			},
			4,
		},
		{
			"Include bot accounts and deleted accounts with existing team id and view restrictions not allowing current team",
			model.UserCountOptions{
				IncludeBotAccounts: true,
				IncludeDeleted:     true,
				TeamId:             teamId,
				ViewRestrictions:   &model.ViewUsersRestrictions{Teams: []string{model.NewId()}},
			},
			0,
		},
		{
			"Filter by system admins only",
			model.UserCountOptions{
				TeamId: teamId,
				Roles:  []string{model.SYSTEM_ADMIN_ROLE_ID},
			},
			1,
		},
		{
			"Filter by system users only",
			model.UserCountOptions{
				TeamId: teamId,
				Roles:  []string{model.SYSTEM_USER_ROLE_ID},
			},
			2,
		},
		{
			"Filter by system guests only",
			model.UserCountOptions{
				TeamId: teamId,
				Roles:  []string{model.SYSTEM_GUEST_ROLE_ID},
			},
			1,
		},
		{
			"Filter by system admins and system users",
			model.UserCountOptions{
				TeamId: teamId,
				Roles:  []string{model.SYSTEM_ADMIN_ROLE_ID, model.SYSTEM_USER_ROLE_ID},
			},
			3,
		},
		{
			"Filter by system admins, system user and system guests",
			model.UserCountOptions{
				TeamId: teamId,
				Roles:  []string{model.SYSTEM_ADMIN_ROLE_ID, model.SYSTEM_USER_ROLE_ID, model.SYSTEM_GUEST_ROLE_ID},
			},
			4,
		},
		{
			"Filter by team admins",
			model.UserCountOptions{
				TeamId:    teamId,
				TeamRoles: []string{model.TEAM_ADMIN_ROLE_ID},
			},
			1,
		},
		{
			"Filter by team members",
			model.UserCountOptions{
				TeamId:    teamId,
				TeamRoles: []string{model.TEAM_USER_ROLE_ID},
			},
			1,
		},
		{
			"Filter by team guests",
			model.UserCountOptions{
				TeamId:    teamId,
				TeamRoles: []string{model.TEAM_GUEST_ROLE_ID},
			},
			1,
		},
		{
			"Filter by team guests and any system role",
			model.UserCountOptions{
				TeamId:    teamId,
				TeamRoles: []string{model.TEAM_GUEST_ROLE_ID},
				Roles:     []string{model.SYSTEM_ADMIN_ROLE_ID},
			},
			2,
		},
		{
			"Filter by channel members",
			model.UserCountOptions{
				ChannelId:    channelId,
				ChannelRoles: []string{model.CHANNEL_USER_ROLE_ID},
			},
			1,
		},
		{
			"Filter by channel members and system admins",
			model.UserCountOptions{
				ChannelId:    channelId,
				Roles:        []string{model.SYSTEM_ADMIN_ROLE_ID},
				ChannelRoles: []string{model.CHANNEL_USER_ROLE_ID},
			},
			2,
		},
		{
			"Filter by channel members and system admins and channel admins",
			model.UserCountOptions{
				ChannelId:    channelId,
				Roles:        []string{model.SYSTEM_ADMIN_ROLE_ID},
				ChannelRoles: []string{model.CHANNEL_USER_ROLE_ID, model.CHANNEL_ADMIN_ROLE_ID},
			},
			3,
		},
		{
			"Filter by channel guests",
			model.UserCountOptions{
				ChannelId:    channelId,
				ChannelRoles: []string{model.CHANNEL_GUEST_ROLE_ID},
			},
			1,
		},
		{
			"Filter by channel guests and any system role",
			model.UserCountOptions{
				ChannelId:    channelId,
				ChannelRoles: []string{model.CHANNEL_GUEST_ROLE_ID},
				Roles:        []string{model.SYSTEM_ADMIN_ROLE_ID},
			},
			2,
		},
	}
	for _, testCase := range testCases {
		s.T().Run(testCase.Description, func(t *testing.T) {
			count, err := s.Store().User().Count(testCase.Options)
			s.Require().Nil(err)
			s.Require().Equal(testCase.Expected, count)
		})
	}
}

func (s *UserStoreTS) TestSave() {
	teamId := model.NewId()
	maxUsersPerTeam := 50

	u1 := model.User{
		Email:    MakeEmail(),
		Username: model.NewId(),
	}

	_, err := s.Store().User().Save(&u1)
	s.Require().Nil(err, "couldn't save user")

	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u1.Id)) }()

	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, maxUsersPerTeam)
	s.Require().Nil(nErr)

	_, err = s.Store().User().Save(&u1)
	s.Require().NotNil(err, "shouldn't be able to update user from save")

	u2 := model.User{
		Email:    u1.Email,
		Username: model.NewId(),
	}
	_, err = s.Store().User().Save(&u2)
	s.Require().NotNil(err, "should be unique email")

	u2.Email = MakeEmail()
	u2.Username = u1.Username
	_, err = s.Store().User().Save(&u2)
	s.Require().NotNil(err, "should be unique username")

	u2.Username = ""
	_, err = s.Store().User().Save(&u2)
	s.Require().NotNil(err, "should be unique username")

	for i := 0; i < 49; i++ {
		u := model.User{
			Email:    MakeEmail(),
			Username: model.NewId(),
		}
		_, err = s.Store().User().Save(&u)
		s.Require().Nil(err, "couldn't save item")

		defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u.Id)) }()

		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u.Id}, maxUsersPerTeam)
		s.Require().Nil(nErr)
	}

	u2.Id = ""
	u2.Email = MakeEmail()
	u2.Username = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err, "couldn't save item")

	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u2.Id)) }()

	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, maxUsersPerTeam)
	s.Require().NotNil(nErr, "should be the limit")
}

func (s *UserStoreTS) TestUpdate() {
	u1 := &model.User{
		Email: MakeEmail(),
	}
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u1.Id)) }()
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{
		Email:       MakeEmail(),
		AuthService: "ldap",
	}
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u2.Id)) }()
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	_, err = s.Store().User().Update(u1, false)
	s.Require().Nil(err)

	missing := &model.User{}
	_, err = s.Store().User().Update(missing, false)
	s.Require().NotNil(err, "Update should have failed because of missing key")

	newId := &model.User{
		Id: model.NewId(),
	}
	_, err = s.Store().User().Update(newId, false)
	s.Require().NotNil(err, "Update should have failed because id change")

	u2.Email = MakeEmail()
	_, err = s.Store().User().Update(u2, false)
	s.Require().NotNil(err, "Update should have failed because you can't modify AD/LDAP fields")

	u3 := &model.User{
		Email:       MakeEmail(),
		AuthService: "gitlab",
	}
	oldEmail := u3.Email
	_, err = s.Store().User().Save(u3)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u3.Id)) }()
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u3.Id}, -1)
	s.Require().Nil(nErr)

	u3.Email = MakeEmail()
	userUpdate, err := s.Store().User().Update(u3, false)
	s.Require().Nil(err, "Update should not have failed")
	s.Assert().Equal(oldEmail, userUpdate.New.Email, "Email should not have been updated as the update is not trusted")

	u3.Email = MakeEmail()
	userUpdate, err = s.Store().User().Update(u3, true)
	s.Require().Nil(err, "Update should not have failed")
	s.Assert().NotEqual(oldEmail, userUpdate.New.Email, "Email should have been updated as the update is trusted")

	err = s.Store().User().UpdateLastPictureUpdate(u1.Id)
	s.Require().Nil(err, "Update should not have failed")
}

func (s *UserStoreTS) TestResetLastPictureUpdate() {

	u1 := &model.User{}
	u1.Email = MakeEmail()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u1.Id)) }()
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	err = s.Store().User().UpdateLastPictureUpdate(u1.Id)
	s.Require().Nil(err)

	user, err := s.Store().User().Get(u1.Id)
	s.Require().Nil(err)

	s.Assert().NotZero(user.LastPictureUpdate)
	s.Assert().NotZero(user.UpdateAt)

	// Ensure update at timestamp changes
	time.Sleep(time.Millisecond)

	err = s.Store().User().ResetLastPictureUpdate(u1.Id)
	s.Require().Nil(err)

	s.Store().User().InvalidateProfileCacheForUser(u1.Id)

	user2, err := s.Store().User().Get(u1.Id)
	s.Require().Nil(err)

	s.Assert().True(user2.UpdateAt > user.UpdateAt)
	s.Assert().Zero(user2.LastPictureUpdate)
}

func (s *UserStoreTS) TestUpdatePassword() {
	teamId := model.NewId()

	u1 := &model.User{}
	u1.Email = MakeEmail()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(u1.Id)) }()
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	hashedPassword := model.HashPassword("newpwd")

	err = s.Store().User().UpdatePassword(u1.Id, hashedPassword)
	s.Require().Nil(err)

	user, err := s.Store().User().GetByEmail(u1.Email)
	s.Require().Nil(err)
	s.Require().Equal(user.Password, hashedPassword, "Password was not updated correctly")
}

type UserStoreGetAllProfilesTS struct {
	suite.Suite
	StoreTestSuite

	u1 *model.User
	u2 *model.User
	u3 *model.User
	u4 *model.User
	u5 *model.User
	u6 *model.User
	u7 *model.User
}

func TestUserStoreGetAllProfilesTS(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &UserStoreGetAllProfilesTS{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *UserStoreGetAllProfilesTS) SetupSuite() {
	// Clean all user first
	users, err := s.Store().User().GetAll()
	s.Require().Nil(err, "failed cleaning up test users")

	for _, u := range users {
		err := s.Store().User().PermanentDelete(u.Id)
		s.Require().Nil(err, "failed cleaning up test user %s", u.Username)
	}

	u1, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u1" + model.NewId(),
	})
	s.Require().Nil(err)
	s.u1 = u1

	u2, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u2" + model.NewId(),
	})
	s.Require().Nil(err)
	s.u2 = u2

	u3, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u3" + model.NewId(),
	})
	s.Require().Nil(err)
	s.u3 = u3

	//TODO: Open this
	// _, nErr := s.Store().Bot().Save(&model.Bot{
	// 	UserId:   u3.Id,
	// 	Username: u3.Username,
	// 	OwnerId:  u1.Id,
	// })
	// s.Require().Nil(nErr)
	// u3.IsBot = true

	u4, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u4" + model.NewId(),
		Roles:    "system_user some-other-role",
	})
	s.Require().Nil(err)
	s.u4 = u4

	u5, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u5" + model.NewId(),
		Roles:    "system_admin",
	})
	s.Require().Nil(err)
	s.u5 = u5

	u6, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u6" + model.NewId(),
		DeleteAt: model.GetMillis(),
		Roles:    "system_admin",
	})
	s.Require().Nil(err)
	s.u6 = u6

	u7, err := s.Store().User().Save(&model.User{
		Email:    storetest.MakeEmail(),
		Username: "u7" + model.NewId(),
		DeleteAt: model.GetMillis(),
	})
	s.Require().Nil(err)
	s.u7 = u7
}

func (s *UserStoreGetAllProfilesTS) TearDownSuite() {
	s.Require().Nil(s.Store().User().PermanentDelete(s.u1.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u2.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u3.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u4.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u5.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u6.Id))
	s.Require().Nil(s.Store().User().PermanentDelete(s.u7.Id))
}

func (s *UserStoreGetAllProfilesTS) TestGetOffset0Limit100() {
	options := &model.UserGetOptions{Page: 0, PerPage: 100}
	actual, err := s.Store().User().GetAllProfiles(options)
	s.Require().Nil(err)

	s.Require().Equal([]*model.User{
		sanitized(s.u1),
		sanitized(s.u2),
		sanitized(s.u3),
		sanitized(s.u4),
		sanitized(s.u5),
		sanitized(s.u6),
		sanitized(s.u7),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestGetOffset0Limit1() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:    0,
		PerPage: 1,
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u1),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestGetAll() {
	actual, err := s.Store().User().GetAll()
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		s.u1,
		s.u2,
		s.u3,
		s.u4,
		s.u5,
		s.u6,
		s.u7,
	}, actual)
}

//   //TODO: Open
// 	// s.T().Run("etag changes for all after user creation", func(t *testing.T) {
// 	// 	etag := s.Store().User().GetEtagForAllProfiles()

// 	// 	uNew := &model.User{}
// 	// 	uNew.Email = storetest.MakeEmail()
// 	// 	_, err := s.Store().User().Save(uNew)
// 	// 	s.Require().Nil(err)
// 	// 	defer func() { s.Require().Nil(s.Store().User().PermanentDelete(uNew.Id)) }()

// 	// 	updatedEtag := s.Store().User().GetEtagForAllProfiles()
// 	// 	s.Require().NotEqual(t, etag, updatedEtag)
// 	// })

func (s *UserStoreGetAllProfilesTS) TestFilterToSystemAdminRole() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:    0,
		PerPage: 10,
		Role:    "system_admin",
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u5),
		sanitized(s.u6),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestFilterToSystemAdminRoleInactive() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:     0,
		PerPage:  10,
		Role:     "system_admin",
		Inactive: true,
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u6),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestFilterToInactive() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:     0,
		PerPage:  10,
		Inactive: true,
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u6),
		sanitized(s.u7),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestFilterToActive() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:    0,
		PerPage: 10,
		Active:  true,
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u1),
		sanitized(s.u2),
		sanitized(s.u3),
		sanitized(s.u4),
		sanitized(s.u5),
	}, actual)
}

func (s *UserStoreGetAllProfilesTS) TestTryToFilterToActiveAndInactive() {
	actual, err := s.Store().User().GetAllProfiles(&model.UserGetOptions{
		Page:     0,
		PerPage:  10,
		Inactive: true,
		Active:   true,
	})
	s.Require().Nil(err)
	s.Require().Equal([]*model.User{
		sanitized(s.u6),
		sanitized(s.u7),
	}, actual)
}
