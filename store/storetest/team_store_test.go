package storetest

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TeamStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func TestTeamStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &TeamStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

// Each user should have a mention count of exactly 1 in the DB at this point.
func (s *TeamStoreTestSuite) TestGetMembersOrderByUserId() {
	teamId1 := model.NewId()
	teamId2 := model.NewId()

	m1 := &model.TeamMember{TeamId: teamId1, UserId: "55555555555555555555555555"}
	m2 := &model.TeamMember{TeamId: teamId1, UserId: "11111111111111111111111111"}
	m3 := &model.TeamMember{TeamId: teamId1, UserId: "33333333333333333333333333"}
	m4 := &model.TeamMember{TeamId: teamId1, UserId: "22222222222222222222222222"}
	m5 := &model.TeamMember{TeamId: teamId1, UserId: "44444444444444444444444444"}
	m6 := &model.TeamMember{TeamId: teamId2, UserId: "00000000000000000000000000"}

	_, nErr := s.Store().Team().SaveMultipleMembers([]*model.TeamMember{m1, m2, m3, m4, m5, m6}, -1)
	s.Require().Nil(nErr)

	// Gets users ordered by UserId
	ms, err := s.Store().Team().GetMembers(teamId1, 0, 100, nil)
	s.Require().Nil(err)
	s.Len(ms, 5)
	s.Equal("11111111111111111111111111", ms[0].UserId)
	s.Equal("22222222222222222222222222", ms[1].UserId)
	s.Equal("33333333333333333333333333", ms[2].UserId)
	s.Equal("44444444444444444444444444", ms[3].UserId)
	s.Equal("55555555555555555555555555", ms[4].UserId)
}

func (s *TeamStoreTestSuite) TestGetMembersOrderByUsernameAndExcludeDeletedMembers() {
	teamId1 := model.NewId()
	teamId2 := model.NewId()
	u1 := &model.User{Username: "a", Email: MakeEmail(), DeleteAt: int64(1)}
	u2 := &model.User{Username: "c", Email: MakeEmail()}
	u3 := &model.User{Username: "b", Email: MakeEmail(), DeleteAt: int64(1)}
	u4 := &model.User{Username: "f", Email: MakeEmail()}
	u5 := &model.User{Username: "e", Email: MakeEmail(), DeleteAt: int64(1)}
	u6 := &model.User{Username: "d", Email: MakeEmail()}
	u1, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	u2, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	u3, err = s.Store().User().Save(u3)
	s.Require().Nil(err)
	u4, err = s.Store().User().Save(u4)
	s.Require().Nil(err)
	u5, err = s.Store().User().Save(u5)
	s.Require().Nil(err)
	u6, err = s.Store().User().Save(u6)
	s.Require().Nil(err)
	m1 := &model.TeamMember{TeamId: teamId1, UserId: u1.Id}
	m2 := &model.TeamMember{TeamId: teamId1, UserId: u2.Id}
	m3 := &model.TeamMember{TeamId: teamId1, UserId: u3.Id}
	m4 := &model.TeamMember{TeamId: teamId1, UserId: u4.Id}
	m5 := &model.TeamMember{TeamId: teamId1, UserId: u5.Id}
	m6 := &model.TeamMember{TeamId: teamId2, UserId: u6.Id}
	_, nErr := s.Store().Team().SaveMultipleMembers([]*model.TeamMember{m1, m2, m3, m4, m5, m6}, -1)
	s.Require().Nil(nErr)
	// Gets users ordered by UserName
	ms, err := s.Store().Team().GetMembers(teamId1, 0, 100, &model.TeamMembersGetOptions{Sort: model.USERNAME})
	s.Require().Nil(err)
	s.Len(ms, 5)
	s.Equal(u1.Id, ms[0].UserId)
	s.Equal(u3.Id, ms[1].UserId)
	s.Equal(u2.Id, ms[2].UserId)
	s.Equal(u5.Id, ms[3].UserId)
	s.Equal(u4.Id, ms[4].UserId)
	// Gets users ordered by UserName and excludes deleted members
	ms, err = s.Store().Team().GetMembers(teamId1, 0, 100, &model.TeamMembersGetOptions{Sort: model.USERNAME, ExcludeDeletedUsers: true})
	s.Require().Nil(err)
	s.Len(ms, 2)
	s.Equal(u2.Id, ms[0].UserId)
	s.Equal(u4.Id, ms[1].UserId)
}

func (s *TeamStoreTestSuite) TestGetMembersExcludedDeletedUsers() {
	teamId1 := model.NewId()
	teamId2 := model.NewId()

	u1 := &model.User{Email: MakeEmail()}
	u2 := &model.User{Email: MakeEmail(), DeleteAt: int64(1)}
	u3 := &model.User{Email: MakeEmail()}
	u4 := &model.User{Email: MakeEmail(), DeleteAt: int64(3)}
	u5 := &model.User{Email: MakeEmail()}
	u6 := &model.User{Email: MakeEmail(), DeleteAt: int64(5)}

	u1, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	u2, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	u3, err = s.Store().User().Save(u3)
	s.Require().Nil(err)
	u4, err = s.Store().User().Save(u4)
	s.Require().Nil(err)
	u5, err = s.Store().User().Save(u5)
	s.Require().Nil(err)
	u6, err = s.Store().User().Save(u6)
	s.Require().Nil(err)

	m1 := &model.TeamMember{TeamId: teamId1, UserId: u1.Id}
	m2 := &model.TeamMember{TeamId: teamId1, UserId: u2.Id}
	m3 := &model.TeamMember{TeamId: teamId1, UserId: u3.Id}
	m4 := &model.TeamMember{TeamId: teamId1, UserId: u4.Id}
	m5 := &model.TeamMember{TeamId: teamId1, UserId: u5.Id}
	m6 := &model.TeamMember{TeamId: teamId2, UserId: u6.Id}

	t1, nErr := s.Store().Team().SaveMember(m1, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Team().SaveMember(m2, -1)
	s.Require().Nil(nErr)
	t3, nErr := s.Store().Team().SaveMember(m3, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Team().SaveMember(m4, -1)
	s.Require().Nil(nErr)
	t5, nErr := s.Store().Team().SaveMember(m5, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Team().SaveMember(m6, -1)
	s.Require().Nil(nErr)

	// Gets users ordered by UserName
	ms, err := s.Store().Team().GetMembers(teamId1, 0, 100, &model.TeamMembersGetOptions{ExcludeDeletedUsers: true})
	s.Require().Nil(err)
	s.Len(ms, 3)
	s.Require().ElementsMatch(ms, [3]*model.TeamMember{t1, t3, t5})
}
