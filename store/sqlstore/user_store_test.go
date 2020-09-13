package sqlstore

import (
	"github.com/masterhung0112/go_server/model"
	"github.com/stretchr/testify/suite"
	"testing"

	"github.com/masterhung0112/go_server/store/storetest"
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
