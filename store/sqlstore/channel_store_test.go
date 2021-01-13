package sqlstore

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/services/timezones"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/stretchr/testify/suite"
)

type ChannelStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func (s *ChannelStoreTestSuite) SetupTest() {
	createDefaultRoles(s.Store())
}

func TestChannelStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &ChannelStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func (s *ChannelStoreTestSuite) cleanupChannels() {
	list, err := s.Store().Channel().GetAllChannels(0, 100000, store.ChannelSearchOpts{IncludeDeleted: true})
	s.Require().Nilf(err, "error cleaning all channels: %v", err)
	for _, channel := range *list {
		err = s.Store().Channel().PermanentDelete(channel.Id)
		s.Assert().NoError(err)
	}
}

func (s *ChannelStoreTestSuite) TestStoreSave() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr, "couldn't save item", nErr)

	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().NotNil(nErr, "shouldn't be able to update from save")

	o1.Id = ""
	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().NotNil(nErr, "should be unique name")

	o1.Id = ""
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT
	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().NotNil(nErr, "should not be able to save direct channel")

	o1 = model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr, "should have saved channel")

	o2 := o1
	o2.Id = ""

	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().NotNil(nErr, "should have failed to save a duplicate channel")
	var cErr *store.ErrConflict
	s.Require().True(errors.As(nErr, &cErr))

	err := s.Store().Channel().Delete(o1.Id, 100)
	s.Require().Nil(err, "should have deleted channel")

	o2.Id = ""
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().NotNil(nErr, "should have failed to save a duplicate of an archived channel")
	s.Require().True(errors.As(nErr, &cErr))
}

func (s *ChannelStoreTestSuite) TestStoreSaveDirectChannel() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	_, nErr = s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)
	s.Require().Nil(nErr, "couldn't save direct channel", nErr)

	members, nErr := s.Store().Channel().GetMembers(o1.Id, 0, 100)
	s.Require().Nil(nErr)
	s.Require().Len(*members, 2, "should have saved 2 members")

	_, nErr = s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)
	s.Require().NotNil(nErr, "shoudn't be a able to update from save")

	// Attempt to save a direct channel that already exists
	o1a := model.Channel{
		TeamId:      o1.TeamId,
		DisplayName: o1.DisplayName,
		Name:        o1.Name,
		Type:        o1.Type,
	}

	returnedChannel, nErr := s.Store().Channel().SaveDirectChannel(&o1a, &m1, &m2)
	s.Require().NotNil(nErr, "should've failed to save a duplicate direct channel")
	var cErr *store.ErrConflict
	s.Require().Truef(errors.As(nErr, &cErr), "should've returned CHANNEL_EXISTS_ERROR")
	s.Require().Equal(o1.Id, returnedChannel.Id, "should've failed to save a duplicate direct channel")

	// Attempt to save a non-direct channel
	o1.Id = ""
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)
	s.Require().NotNil(nErr, "Should not be able to save non-direct channel")

	// Save yourself Direct Message
	o1.Id = ""
	o1.DisplayName = "Myself"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT
	_, nErr = s.Store().Channel().SaveDirectChannel(&o1, &m1, &m1)
	s.Require().Nil(nErr, "couldn't save direct channel", nErr)

	members, nErr = s.Store().Channel().GetMembers(o1.Id, 0, 100)
	s.Require().Nil(nErr)
	s.Require().Len(*members, 1, "should have saved just 1 member")

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreCreateDirectChannel() {
	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	c1, nErr := s.Store().Channel().CreateDirectChannel(u1, u2)
	s.Require().Nil(nErr, "couldn't create direct channel", nErr)
	defer func() {
		s.Store().Channel().PermanentDeleteMembersByChannel(c1.Id)
		s.Store().Channel().PermanentDelete(c1.Id)
	}()

	members, nErr := s.Store().Channel().GetMembers(c1.Id, 0, 100)
	s.Require().Nil(nErr)
	s.Require().Len(*members, 2, "should have saved 2 members")
}

func (s *ChannelStoreTestSuite) TestStoreUpdate() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Name"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN

	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	time.Sleep(100 * time.Millisecond)

	_, err := s.Store().Channel().Update(&o1)
	s.Require().Nil(err, err)

	o1.DeleteAt = 100
	_, err = s.Store().Channel().Update(&o1)
	s.Require().NotNil(err, "update should have failed because channel is archived")

	o1.DeleteAt = 0
	o1.Id = "missing"
	_, err = s.Store().Channel().Update(&o1)
	s.Require().NotNil(err, "Update should have failed because of missing key")

	o2.Name = o1.Name
	_, err = s.Store().Channel().Update(&o2)
	s.Require().NotNil(err, "update should have failed because of existing name")
}

func (s *ChannelStoreTestSuite) TestGetChannelUnread() {
	teamId1 := model.NewId()
	teamId2 := model.NewId()

	uid := model.NewId()
	m1 := &model.TeamMember{TeamId: teamId1, UserId: uid}
	m2 := &model.TeamMember{TeamId: teamId2, UserId: uid}
	_, nErr := s.Store().Team().SaveMember(m1, -1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Team().SaveMember(m2, -1)
	s.Require().Nil(nErr)
	notifyPropsModel := model.GetDefaultChannelNotifyProps()

	// Setup Channel 1
	c1 := &model.Channel{TeamId: m1.TeamId, Name: model.NewId(), DisplayName: "Downtown", Type: model.CHANNEL_OPEN, TotalMsgCount: 100}
	_, nErr = s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	cm1 := &model.ChannelMember{ChannelId: c1.Id, UserId: m1.UserId, NotifyProps: notifyPropsModel, MsgCount: 90}
	_, err := s.Store().Channel().SaveMember(cm1)
	s.Require().Nil(err)

	// Setup Channel 2
	c2 := &model.Channel{TeamId: m2.TeamId, Name: model.NewId(), DisplayName: "Cultural", Type: model.CHANNEL_OPEN, TotalMsgCount: 100}
	_, nErr = s.Store().Channel().Save(c2, -1)
	s.Require().Nil(nErr)

	cm2 := &model.ChannelMember{ChannelId: c2.Id, UserId: m2.UserId, NotifyProps: notifyPropsModel, MsgCount: 90, MentionCount: 5}
	_, err = s.Store().Channel().SaveMember(cm2)
	s.Require().Nil(err)

	// Check for Channel 1
	ch, nErr := s.Store().Channel().GetChannelUnread(c1.Id, uid)

	s.Require().Nil(nErr, nErr)
	s.Require().Equal(c1.Id, ch.ChannelId, "Wrong channel id")
	s.Require().Equal(teamId1, ch.TeamId, "Wrong team id for channel 1")
	s.Require().NotNil(ch.NotifyProps, "wrong props for channel 1")
	s.Require().EqualValues(0, ch.MentionCount, "wrong MentionCount for channel 1")
	s.Require().EqualValues(10, ch.MsgCount, "wrong MsgCount for channel 1")

	// Check for Channel 2
	ch2, nErr := s.Store().Channel().GetChannelUnread(c2.Id, uid)

	s.Require().Nil(nErr, nErr)
	s.Require().Equal(c2.Id, ch2.ChannelId, "Wrong channel id")
	s.Require().Equal(teamId2, ch2.TeamId, "Wrong team id")
	s.Require().EqualValues(5, ch2.MentionCount, "wrong MentionCount for channel 2")
	s.Require().EqualValues(10, ch2.MsgCount, "wrong MsgCount for channel 2")
}

func (s *ChannelStoreTestSuite) TestStoreGet() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	c1 := &model.Channel{}
	c1, err := s.Store().Channel().Get(o1.Id, false)
	s.Require().Nil(err, err)
	s.Require().Equal(o1.ToJson(), c1.ToJson(), "invalid returned channel")

	_, err = s.Store().Channel().Get("", false)
	s.Require().NotNil(err, "missing id should have failed")

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err = s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Direct Name"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_DIRECT

	m1 := model.ChannelMember{}
	m1.ChannelId = o2.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	_, nErr = s.Store().Channel().SaveDirectChannel(&o2, &m1, &m2)
	s.Require().Nil(nErr)

	c2, err := s.Store().Channel().Get(o2.Id, false)
	s.Require().Nil(err, err)
	s.Require().Equal(o2.ToJson(), c2.ToJson(), "invalid returned channel")

	c4, err := s.Store().Channel().Get(o2.Id, true)
	s.Require().Nil(err, err)
	s.Require().Equal(o2.ToJson(), c4.ToJson(), "invalid returned channel")

	channels, chanErr := s.Store().Channel().GetAll(o1.TeamId)
	s.Require().Nil(chanErr, chanErr)
	s.Require().Greater(len(channels), 0, "too little")

	channelsTeam, err := s.Store().Channel().GetTeamChannels(o1.TeamId)
	s.Require().Nil(err, err)
	s.Require().Greater(len(*channelsTeam), 0, "too little")

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreGetChannelsByIds() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "aa" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Direct Name"
	o2.Name = "bb" + model.NewId() + "b"
	o2.Type = model.CHANNEL_DIRECT

	o3 := model.Channel{}
	o3.TeamId = model.NewId()
	o3.DisplayName = "Deleted channel"
	o3.Name = "cc" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)
	nErr = s.Store().Channel().Delete(o3.Id, 123)
	s.Require().Nil(nErr)
	o3.DeleteAt = 123
	o3.UpdateAt = 123

	m1 := model.ChannelMember{}
	m1.ChannelId = o2.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	_, nErr = s.Store().Channel().SaveDirectChannel(&o2, &m1, &m2)
	s.Require().Nil(nErr)

	s.T().Run("Get 2 existing channels", func(t *testing.T) {
		r1, err := s.Store().Channel().GetChannelsByIds([]string{o1.Id, o2.Id}, false)
		s.Require().Nil(err, err)
		s.Require().Len(r1, 2, "invalid returned channels, exepected 2 and got "+strconv.Itoa(len(r1)))
		s.Require().Equal(o1.ToJson(), r1[0].ToJson())
		s.Require().Equal(o2.ToJson(), r1[1].ToJson())
	})

	s.T().Run("Get 1 existing and 1 not existing channel", func(t *testing.T) {
		nonexistentId := "abcd1234"
		r2, err := s.Store().Channel().GetChannelsByIds([]string{o1.Id, nonexistentId}, false)
		s.Require().Nil(err, err)
		s.Require().Len(r2, 1, "invalid returned channels, expected 1 and got "+strconv.Itoa(len(r2)))
		s.Require().Equal(o1.ToJson(), r2[0].ToJson(), "invalid returned channel")
	})

	s.T().Run("Get 2 existing and 1 deleted channel", func(t *testing.T) {
		r1, err := s.Store().Channel().GetChannelsByIds([]string{o1.Id, o2.Id, o3.Id}, true)
		s.Require().Nil(err, err)
		s.Require().Len(r1, 3, "invalid returned channels, exepected 3 and got "+strconv.Itoa(len(r1)))
		s.Require().Equal(o1.ToJson(), r1[0].ToJson())
		s.Require().Equal(o2.ToJson(), r1[1].ToJson())
		s.Require().Equal(o3.ToJson(), r1[2].ToJson())
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetForPost() {

	ch := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	o1, nErr := s.Store().Channel().Save(ch, -1)
	s.Require().Nil(nErr)

	p1, err := s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
	})
	s.Require().Nil(err)

	channel, chanErr := s.Store().Channel().GetForPost(p1.Id)
	s.Require().Nil(chanErr, chanErr)
	s.Require().Equal(o1.Id, channel.Id, "incorrect channel returned")
}

func (s *ChannelStoreTestSuite) TestStoreRestore() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	err := s.Store().Channel().Delete(o1.Id, model.GetMillis())
	s.Require().Nil(err, err)

	c, _ := s.Store().Channel().Get(o1.Id, false)
	s.Require().NotEqual(0, c.DeleteAt, "should have been deleted")

	err = s.Store().Channel().Restore(o1.Id, model.GetMillis())
	s.Require().Nil(err, err)

	c, _ = s.Store().Channel().Get(o1.Id, false)
	s.Require().EqualValues(0, c.DeleteAt, "should have been restored")
}

func (s *ChannelStoreTestSuite) TestStoreDelete() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "Channel3"
	o3.Name = "zz" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	o4 := model.Channel{}
	o4.TeamId = o1.TeamId
	o4.DisplayName = "Channel4"
	o4.Name = "zz" + model.NewId() + "b"
	o4.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	nErr = s.Store().Channel().Delete(o1.Id, model.GetMillis())
	s.Require().Nil(nErr, nErr)

	c, _ := s.Store().Channel().Get(o1.Id, false)
	s.Require().NotEqual(0, c.DeleteAt, "should have been deleted")

	nErr = s.Store().Channel().Delete(o3.Id, model.GetMillis())
	s.Require().Nil(nErr, nErr)

	list, nErr := s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, false, 0)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 1, "invalid number of channels")

	list, nErr = s.Store().Channel().GetMoreChannels(o1.TeamId, m1.UserId, 0, 100)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 1, "invalid number of channels")

	cresult := s.Store().Channel().PermanentDelete(o2.Id)
	s.Require().Nil(cresult)

	list, nErr = s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, false, 0)
	if s.Assert().NotNil(nErr) {
		var nfErr *store.ErrNotFound
		s.Require().True(errors.As(nErr, &nfErr))
	} else {
		s.Require().Equal(&model.ChannelList{}, list)
	}

	nErr = s.Store().Channel().PermanentDeleteByTeam(o1.TeamId)
	s.Require().Nil(nErr, nErr)
}

func (s *ChannelStoreTestSuite) TestStoreGetByName() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	result, err := s.Store().Channel().GetByName(o1.TeamId, o1.Name, true)
	s.Require().Nil(err)
	s.Require().Equal(o1.ToJson(), result.ToJson(), "invalid returned channel")

	channelID := result.Id

	result, err = s.Store().Channel().GetByName(o1.TeamId, "", true)
	s.Require().NotNil(err, "Missing id should have failed")

	result, err = s.Store().Channel().GetByName(o1.TeamId, o1.Name, false)
	s.Require().Nil(err)
	s.Require().Equal(o1.ToJson(), result.ToJson(), "invalid returned channel")

	result, err = s.Store().Channel().GetByName(o1.TeamId, "", false)
	s.Require().NotNil(err, "Missing id should have failed")

	nErr = s.Store().Channel().Delete(channelID, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")

	result, err = s.Store().Channel().GetByName(o1.TeamId, o1.Name, false)
	s.Require().NotNil(err, "Deleted channel should not be returned by GetByName()")
}

func (s *ChannelStoreTestSuite) TestStoreGetByNames() {
	o1 := model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{
		TeamId:      o1.TeamId,
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	for index, tc := range []struct {
		TeamId      string
		Names       []string
		ExpectedIds []string
	}{
		{o1.TeamId, []string{o1.Name}, []string{o1.Id}},
		{o1.TeamId, []string{o1.Name, o2.Name}, []string{o1.Id, o2.Id}},
		{o1.TeamId, nil, nil},
		{o1.TeamId, []string{"foo"}, nil},
		{o1.TeamId, []string{o1.Name, "foo", o2.Name, o2.Name}, []string{o1.Id, o2.Id}},
		{"", []string{o1.Name, "foo", o2.Name, o2.Name}, []string{o1.Id, o2.Id}},
		{"asd", []string{o1.Name, "foo", o2.Name, o2.Name}, nil},
	} {
		var channels []*model.Channel
		channels, err := s.Store().Channel().GetByNames(tc.TeamId, tc.Names, true)
		s.Require().Nil(err)
		var ids []string
		for _, channel := range channels {
			ids = append(ids, channel.Id)
		}
		sort.Strings(ids)
		sort.Strings(tc.ExpectedIds)
		s.Assert().Equal(tc.ExpectedIds, ids, "tc %v", index)
	}

	err := s.Store().Channel().Delete(o1.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	err = s.Store().Channel().Delete(o2.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	channels, nErr := s.Store().Channel().GetByNames(o1.TeamId, []string{o1.Name}, false)
	s.Require().Nil(nErr)
	s.Assert().Empty(channels)
}

func (s *ChannelStoreTestSuite) TestStoreGetDeletedByName() {
	o1 := &model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(o1, -1)
	s.Require().Nil(nErr)

	now := model.GetMillis()
	err := s.Store().Channel().Delete(o1.Id, now)
	s.Require().Nil(err, "channel should have been deleted")
	o1.DeleteAt = now
	o1.UpdateAt = now

	r1, nErr := s.Store().Channel().GetDeletedByName(o1.TeamId, o1.Name)
	s.Require().Nil(nErr)
	s.Require().Equal(o1, r1)

	_, nErr = s.Store().Channel().GetDeletedByName(o1.TeamId, "")
	s.Require().NotNil(nErr, "missing id should have failed")
}

func (s *ChannelStoreTestSuite) TestStoreGetDeleted() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN

	userId := model.NewId()

	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	err := s.Store().Channel().Delete(o1.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	list, nErr := s.Store().Channel().GetDeleted(o1.TeamId, 0, 100, userId)
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*list, 1, "wrong list")
	s.Require().Equal(o1.Name, (*list)[0].Name, "missing channel")

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	list, nErr = s.Store().Channel().GetDeleted(o1.TeamId, 0, 100, userId)
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*list, 1, "wrong list")

	o3 := model.Channel{}
	o3.TeamId = o1.TeamId
	o3.DisplayName = "Channel3"
	o3.Name = "zz" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN

	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	err = s.Store().Channel().Delete(o3.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	list, nErr = s.Store().Channel().GetDeleted(o1.TeamId, 0, 100, userId)
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*list, 2, "wrong list length")

	list, nErr = s.Store().Channel().GetDeleted(o1.TeamId, 0, 1, userId)
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*list, 1, "wrong list length")

	list, nErr = s.Store().Channel().GetDeleted(o1.TeamId, 1, 1, userId)
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*list, 1, "wrong list length")

}

func (s *ChannelStoreTestSuite) TestMemberStore() {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "NameName"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	c1t1, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t1.ExtraUpdateAt, "ExtraUpdateAt should be 0")

	u1 := model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(&u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	o1 := model.ChannelMember{}
	o1.ChannelId = c1.Id
	o1.UserId = u1.Id
	o1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, nErr = s.Store().Channel().SaveMember(&o1)
	s.Require().Nil(nErr)

	o2 := model.ChannelMember{}
	o2.ChannelId = c1.Id
	o2.UserId = u2.Id
	o2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, nErr = s.Store().Channel().SaveMember(&o2)
	s.Require().Nil(nErr)

	c1t2, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t2.ExtraUpdateAt, "ExtraUpdateAt should be 0")

	count, nErr := s.Store().Channel().GetMemberCount(o1.ChannelId, true)
	s.Require().Nil(nErr)
	s.Require().EqualValues(2, count, "should have saved 2 members")

	count, nErr = s.Store().Channel().GetMemberCount(o1.ChannelId, true)
	s.Require().Nil(nErr)
	s.Require().EqualValues(2, count, "should have saved 2 members")
	s.Require().EqualValues(
		2,
		s.Store().Channel().GetMemberCountFromCache(o1.ChannelId),
		"should have saved 2 members")

	s.Require().EqualValues(
		0,
		s.Store().Channel().GetMemberCountFromCache("junk"),
		"should have saved 0 members")

	count, nErr = s.Store().Channel().GetMemberCount(o1.ChannelId, false)
	s.Require().Nil(nErr)
	s.Require().EqualValues(2, count, "should have saved 2 members")

	nErr = s.Store().Channel().RemoveMember(o2.ChannelId, o2.UserId)
	s.Require().Nil(nErr)

	count, nErr = s.Store().Channel().GetMemberCount(o1.ChannelId, false)
	s.Require().Nil(nErr)
	s.Require().EqualValues(1, count, "should have removed 1 member")

	c1t3, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t3.ExtraUpdateAt, "ExtraUpdateAt should be 0")

	member, _ := s.Store().Channel().GetMember(o1.ChannelId, o1.UserId)
	s.Require().Equal(o1.ChannelId, member.ChannelId, "should have go member")

	_, nErr = s.Store().Channel().SaveMember(&o1)
	s.Require().NotNil(nErr, "should have been a duplicate")

	c1t4, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t4.ExtraUpdateAt, "ExtraUpdateAt should be 0")
}

func (s *ChannelStoreTestSuite) TestSaveMember() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()

	s.T().Run("not valid channel member", func(t *testing.T) {
		member := &model.ChannelMember{ChannelId: "wrong", UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMember(member)
		s.Require().NotNil(nErr)
		var appErr *model.AppError
		s.Require().True(errors.As(nErr, &appErr))
		s.Require().Equal("model.channel_member.is_valid.channel_id.app_error", appErr.Id)
	})

	s.T().Run("duplicated entries should fail", func(t *testing.T) {
		channelID1 := model.NewId()
		m1 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMember(m1)
		s.Require().Nil(nErr)
		m2 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr = s.Store().Channel().SaveMember(m2)
		s.Require().NotNil(nErr)
		s.Require().IsType(&store.ErrConflict{}, nErr)
	})

	s.T().Run("insert member correctly (in channel without channel scheme and team without scheme)", func(t *testing.T) {
		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr := s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				member, nErr = s.Store().Channel().SaveMember(member)
				s.Require().Nil(nErr)
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert member correctly (in channel without scheme and team with scheme)", func(t *testing.T) {
		ts := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		}
		ts, nErr := s.Store().Scheme().Save(ts)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
			SchemeId:    &ts.Id,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				member, nErr = s.Store().Channel().SaveMember(member)
				s.Require().Nil(nErr)
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert member correctly (in channel with channel scheme)", func(t *testing.T) {
		cs := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		}
		cs, nErr := s.Store().Scheme().Save(cs)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel, nErr := s.Store().Channel().Save(&model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
			SchemeId:    &cs.Id,
		}, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				member, nErr = s.Store().Channel().SaveMember(member)
				s.Require().Nil(nErr)
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})
}

func (s *ChannelStoreTestSuite) TestSaveMultipleMembers() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u2, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()

	s.T().Run("any not valid channel member", func(t *testing.T) {
		m1 := &model.ChannelMember{ChannelId: "wrong", UserId: u1.Id, NotifyProps: defaultNotifyProps}
		m2 := &model.ChannelMember{ChannelId: model.NewId(), UserId: u2.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2})
		s.Require().NotNil(nErr)
		var appErr *model.AppError
		s.Require().True(errors.As(nErr, &appErr))
		s.Require().Equal("model.channel_member.is_valid.channel_id.app_error", appErr.Id)
	})

	s.T().Run("duplicated entries should fail", func(t *testing.T) {
		channelID1 := model.NewId()
		m1 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		m2 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2})
		s.Require().NotNil(nErr)
		s.Require().IsType(&store.ErrConflict{}, nErr)
	})

	s.T().Run("insert members correctly (in channel without channel scheme and team without scheme)", func(t *testing.T) {
		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr := s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				otherMember := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u2.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				var members []*model.ChannelMember
				members, nErr = s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(nErr)
				s.Require().Len(members, 2)
				member = members[0]
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert members correctly (in channel without scheme and team with scheme)", func(t *testing.T) {
		ts := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		}
		ts, nErr := s.Store().Scheme().Save(ts)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
			SchemeId:    &ts.Id,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				otherMember := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u2.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				var members []*model.ChannelMember
				members, nErr = s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(nErr)
				s.Require().Len(members, 2)
				member = members[0]
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert members correctly (in channel with channel scheme)", func(t *testing.T) {
		cs := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		}
		cs, nErr := s.Store().Scheme().Save(cs)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel, nErr := s.Store().Channel().Save(&model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
			SchemeId:    &cs.Id,
		}, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u1.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				otherMember := &model.ChannelMember{
					ChannelId:     channel.Id,
					UserId:        u2.Id,
					SchemeGuest:   tc.SchemeGuest,
					SchemeUser:    tc.SchemeUser,
					SchemeAdmin:   tc.SchemeAdmin,
					ExplicitRoles: tc.ExplicitRoles,
					NotifyProps:   defaultNotifyProps,
				}
				members, err := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(err)
				s.Require().Len(members, 2)
				member = members[0]
				defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
				defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})
}

func (s *ChannelStoreTestSuite) TestUpdateMember() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()

	s.T().Run("not valid channel member", func(t *testing.T) {
		member := &model.ChannelMember{ChannelId: "wrong", UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().UpdateMember(member)
		s.Require().NotNil(nErr)
		var appErr *model.AppError
		s.Require().True(errors.As(nErr, &appErr))
		s.Require().Equal("model.channel_member.is_valid.channel_id.app_error", appErr.Id)
	})

	s.T().Run("insert member correctly (in channel without channel scheme and team without scheme)", func(t *testing.T) {
		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr := s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{
			ChannelId:   channel.Id,
			UserId:      u1.Id,
			NotifyProps: defaultNotifyProps,
		}
		member, nErr = s.Store().Channel().SaveMember(member)
		s.Require().Nil(nErr)

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				member, nErr = s.Store().Channel().UpdateMember(member)
				s.Require().Nil(nErr)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert member correctly (in channel without scheme and team with scheme)", func(t *testing.T) {
		ts := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		}
		ts, nErr := s.Store().Scheme().Save(ts)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
			SchemeId:    &ts.Id,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{
			ChannelId:   channel.Id,
			UserId:      u1.Id,
			NotifyProps: defaultNotifyProps,
		}
		member, nErr = s.Store().Channel().SaveMember(member)
		s.Require().Nil(nErr)

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				member, nErr = s.Store().Channel().UpdateMember(member)
				s.Require().Nil(nErr)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert member correctly (in channel with channel scheme)", func(t *testing.T) {
		cs := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		}
		cs, nErr := s.Store().Scheme().Save(cs)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel, nErr := s.Store().Channel().Save(&model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
			SchemeId:    &cs.Id,
		}, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{
			ChannelId:   channel.Id,
			UserId:      u1.Id,
			NotifyProps: defaultNotifyProps,
		}
		member, nErr = s.Store().Channel().SaveMember(member)
		s.Require().Nil(nErr)

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				member, nErr = s.Store().Channel().UpdateMember(member)
				s.Require().Nil(nErr)
				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})
}

func (s *ChannelStoreTestSuite) TestUpdateMultipleMembers() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u2, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()

	s.T().Run("any not valid channel member", func(t *testing.T) {
		m1 := &model.ChannelMember{ChannelId: "wrong", UserId: u1.Id, NotifyProps: defaultNotifyProps}
		m2 := &model.ChannelMember{ChannelId: model.NewId(), UserId: u2.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2})
		s.Require().NotNil(nErr)
		var appErr *model.AppError
		s.Require().True(errors.As(nErr, &appErr))
		s.Require().Equal("model.channel_member.is_valid.channel_id.app_error", appErr.Id)
	})

	s.T().Run("duplicated entries should fail", func(t *testing.T) {
		channelID1 := model.NewId()
		m1 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		m2 := &model.ChannelMember{ChannelId: channelID1, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2})
		s.Require().NotNil(nErr)
		s.Require().IsType(&store.ErrConflict{}, nErr)
	})

	s.T().Run("insert members correctly (in channel without channel scheme and team without scheme)", func(t *testing.T) {
		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr := s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{ChannelId: channel.Id, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		otherMember := &model.ChannelMember{ChannelId: channel.Id, UserId: u2.Id, NotifyProps: defaultNotifyProps}
		var members []*model.ChannelMember
		members, nErr = s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
		s.Require().Nil(nErr)
		defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
		defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)
		s.Require().Len(members, 2)
		member = members[0]
		otherMember = members[1]

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      "channel_user",
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       "channel_guest",
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       "channel_user channel_admin",
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test channel_user",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test channel_guest",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test channel_user channel_admin",
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				var members []*model.ChannelMember
				members, nErr = s.Store().Channel().UpdateMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(nErr)
				s.Require().Len(members, 2)
				member = members[0]

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert members correctly (in channel without scheme and team with scheme)", func(t *testing.T) {
		ts := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_TEAM,
		}
		ts, nErr := s.Store().Scheme().Save(ts)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
			SchemeId:    &ts.Id,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel := &model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
		}
		channel, nErr = s.Store().Channel().Save(channel, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{ChannelId: channel.Id, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		otherMember := &model.ChannelMember{ChannelId: channel.Id, UserId: u2.Id, NotifyProps: defaultNotifyProps}
		var members []*model.ChannelMember
		members, nErr = s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
		s.Require().Nil(nErr)
		defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
		defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)
		s.Require().Len(members, 2)
		member = members[0]
		otherMember = members[1]

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      ts.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       ts.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + ts.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + ts.DefaultChannelUserRole + " " + ts.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				var members []*model.ChannelMember
				members, nErr = s.Store().Channel().UpdateMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(nErr)
				s.Require().Len(members, 2)
				member = members[0]

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})

	s.T().Run("insert members correctly (in channel with channel scheme)", func(t *testing.T) {
		cs := &model.Scheme{
			Name:        model.NewId(),
			DisplayName: model.NewId(),
			Description: model.NewId(),
			Scope:       model.SCHEME_SCOPE_CHANNEL,
		}
		cs, nErr := s.Store().Scheme().Save(cs)
		s.Require().Nil(nErr)

		team := &model.Team{
			DisplayName: "Name",
			Name:        "zz" + model.NewId(),
			Email:       MakeEmail(),
			Type:        model.TEAM_OPEN,
		}

		team, nErr = s.Store().Team().Save(team)
		s.Require().Nil(nErr)

		channel, nErr := s.Store().Channel().Save(&model.Channel{
			DisplayName: "DisplayName",
			Name:        "z-z-z" + model.NewId() + "b",
			Type:        model.CHANNEL_OPEN,
			TeamId:      team.Id,
			SchemeId:    &cs.Id,
		}, -1)
		s.Require().Nil(nErr)
		defer func() { s.Store().Channel().PermanentDelete(channel.Id) }()

		member := &model.ChannelMember{ChannelId: channel.Id, UserId: u1.Id, NotifyProps: defaultNotifyProps}
		otherMember := &model.ChannelMember{ChannelId: channel.Id, UserId: u2.Id, NotifyProps: defaultNotifyProps}
		members, err := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{member, otherMember})
		s.Require().Nil(err)
		defer s.Store().Channel().RemoveMember(channel.Id, u1.Id)
		defer s.Store().Channel().RemoveMember(channel.Id, u2.Id)
		s.Require().Len(members, 2)
		member = members[0]
		otherMember = members[1]

		testCases := []struct {
			Name                  string
			SchemeGuest           bool
			SchemeUser            bool
			SchemeAdmin           bool
			ExplicitRoles         string
			ExpectedRoles         string
			ExpectedExplicitRoles string
			ExpectedSchemeGuest   bool
			ExpectedSchemeUser    bool
			ExpectedSchemeAdmin   bool
		}{
			{
				Name:               "channel user implicit",
				SchemeUser:         true,
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:               "channel user explicit",
				ExplicitRoles:      "channel_user",
				ExpectedRoles:      cs.DefaultChannelUserRole,
				ExpectedSchemeUser: true,
			},
			{
				Name:                "channel guest implicit",
				SchemeGuest:         true,
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel guest explicit",
				ExplicitRoles:       "channel_guest",
				ExpectedRoles:       cs.DefaultChannelGuestRole,
				ExpectedSchemeGuest: true,
			},
			{
				Name:                "channel admin implicit",
				SchemeUser:          true,
				SchemeAdmin:         true,
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                "channel admin explicit",
				ExplicitRoles:       "channel_user channel_admin",
				ExpectedRoles:       cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedSchemeUser:  true,
				ExpectedSchemeAdmin: true,
			},
			{
				Name:                  "channel user implicit and explicit custom role",
				SchemeUser:            true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel user explicit and explicit custom role",
				ExplicitRoles:         "channel_user test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
			},
			{
				Name:                  "channel guest implicit and explicit custom role",
				SchemeGuest:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel guest explicit and explicit custom role",
				ExplicitRoles:         "channel_guest test",
				ExpectedRoles:         "test " + cs.DefaultChannelGuestRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeGuest:   true,
			},
			{
				Name:                  "channel admin implicit and explicit custom role",
				SchemeUser:            true,
				SchemeAdmin:           true,
				ExplicitRoles:         "test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel admin explicit and explicit custom role",
				ExplicitRoles:         "channel_user channel_admin test",
				ExpectedRoles:         "test " + cs.DefaultChannelUserRole + " " + cs.DefaultChannelAdminRole,
				ExpectedExplicitRoles: "test",
				ExpectedSchemeUser:    true,
				ExpectedSchemeAdmin:   true,
			},
			{
				Name:                  "channel member with only explicit custom roles",
				ExplicitRoles:         "test test2",
				ExpectedRoles:         "test test2",
				ExpectedExplicitRoles: "test test2",
			},
		}

		for _, tc := range testCases {
			s.T().Run(tc.Name, func(t *testing.T) {
				member.SchemeGuest = tc.SchemeGuest
				member.SchemeUser = tc.SchemeUser
				member.SchemeAdmin = tc.SchemeAdmin
				member.ExplicitRoles = tc.ExplicitRoles
				members, err := s.Store().Channel().UpdateMultipleMembers([]*model.ChannelMember{member, otherMember})
				s.Require().Nil(err)
				s.Require().Len(members, 2)
				member = members[0]

				s.Assert().Equal(tc.ExpectedRoles, member.Roles)
				s.Assert().Equal(tc.ExpectedExplicitRoles, member.ExplicitRoles)
				s.Assert().Equal(tc.ExpectedSchemeGuest, member.SchemeGuest)
				s.Assert().Equal(tc.ExpectedSchemeUser, member.SchemeUser)
				s.Assert().Equal(tc.ExpectedSchemeAdmin, member.SchemeAdmin)
			})
		}
	})
}

func (s *ChannelStoreTestSuite) TestRemoveMember() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u2, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u3, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u4, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	channelID := model.NewId()
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()
	m1 := &model.ChannelMember{ChannelId: channelID, UserId: u1.Id, NotifyProps: defaultNotifyProps}
	m2 := &model.ChannelMember{ChannelId: channelID, UserId: u2.Id, NotifyProps: defaultNotifyProps}
	m3 := &model.ChannelMember{ChannelId: channelID, UserId: u3.Id, NotifyProps: defaultNotifyProps}
	m4 := &model.ChannelMember{ChannelId: channelID, UserId: u4.Id, NotifyProps: defaultNotifyProps}
	_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2, m3, m4})
	s.Require().Nil(nErr)

	s.T().Run("remove member from not existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMember("not-existing-channel", u1.Id)
		s.Require().Nil(nErr)
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(4), membersCount)
	})

	s.T().Run("remove not existing member from an existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMember(channelID, model.NewId())
		s.Require().Nil(nErr)
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(4), membersCount)
	})

	s.T().Run("remove existing member from an existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMember(channelID, u1.Id)
		s.Require().Nil(nErr)
		defer s.Store().Channel().SaveMember(m1)
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(3), membersCount)
	})
}

func (s *ChannelStoreTestSuite) TestRemoveMembers() {
	u1, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u2, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u3, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	u4, err := s.Store().User().Save(&model.User{Username: model.NewId(), Email: MakeEmail()})
	s.Require().Nil(err)
	channelID := model.NewId()
	defaultNotifyProps := model.GetDefaultChannelNotifyProps()
	m1 := &model.ChannelMember{ChannelId: channelID, UserId: u1.Id, NotifyProps: defaultNotifyProps}
	m2 := &model.ChannelMember{ChannelId: channelID, UserId: u2.Id, NotifyProps: defaultNotifyProps}
	m3 := &model.ChannelMember{ChannelId: channelID, UserId: u3.Id, NotifyProps: defaultNotifyProps}
	m4 := &model.ChannelMember{ChannelId: channelID, UserId: u4.Id, NotifyProps: defaultNotifyProps}
	_, nErr := s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2, m3, m4})
	s.Require().Nil(nErr)

	s.T().Run("remove members from not existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMembers("not-existing-channel", []string{u1.Id, u2.Id, u3.Id, u4.Id})
		s.Require().Nil(nErr)
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(4), membersCount)
	})

	s.T().Run("remove not existing members from an existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMembers(channelID, []string{model.NewId(), model.NewId()})
		s.Require().Nil(nErr)
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(4), membersCount)
	})

	s.T().Run("remove not existing and not existing members from an existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMembers(channelID, []string{u1.Id, u2.Id, model.NewId(), model.NewId()})
		s.Require().Nil(nErr)
		defer s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2})
		var membersCount int64
		membersCount, nErr = s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(nErr)
		s.Require().Equal(int64(2), membersCount)
	})
	s.T().Run("remove existing members from an existing channel", func(t *testing.T) {
		nErr = s.Store().Channel().RemoveMembers(channelID, []string{u1.Id, u2.Id, u3.Id})
		s.Require().Nil(nErr)
		defer s.Store().Channel().SaveMultipleMembers([]*model.ChannelMember{m1, m2, m3})
		membersCount, err := s.Store().Channel().GetMemberCount(channelID, false)
		s.Require().Nil(err)
		s.Require().Equal(int64(1), membersCount)
	})
}

func (s *ChannelStoreTestSuite) TestDeleteMemberStore() {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "NameName"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	c1t1, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t1.ExtraUpdateAt, "ExtraUpdateAt should be 0")

	u1 := model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(&u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	o1 := model.ChannelMember{}
	o1.ChannelId = c1.Id
	o1.UserId = u1.Id
	o1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, nErr = s.Store().Channel().SaveMember(&o1)
	s.Require().Nil(nErr)

	o2 := model.ChannelMember{}
	o2.ChannelId = c1.Id
	o2.UserId = u2.Id
	o2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, nErr = s.Store().Channel().SaveMember(&o2)
	s.Require().Nil(nErr)

	c1t2, _ := s.Store().Channel().Get(c1.Id, false)
	s.Assert().EqualValues(0, c1t2.ExtraUpdateAt, "ExtraUpdateAt should be 0")

	count, nErr := s.Store().Channel().GetMemberCount(o1.ChannelId, false)
	s.Require().Nil(nErr)
	s.Require().EqualValues(2, count, "should have saved 2 members")

	nErr = s.Store().Channel().PermanentDeleteMembersByUser(o2.UserId)
	s.Require().Nil(nErr)

	count, nErr = s.Store().Channel().GetMemberCount(o1.ChannelId, false)
	s.Require().Nil(nErr)
	s.Require().EqualValues(1, count, "should have removed 1 member")

	nErr = s.Store().Channel().PermanentDeleteMembersByChannel(o1.ChannelId)
	s.Require().Nil(nErr)

	count, nErr = s.Store().Channel().GetMemberCount(o1.ChannelId, false)
	s.Require().Nil(nErr)
	s.Require().EqualValues(0, count, "should have removed all members")
}

func (s *ChannelStoreTestSuite) TestStoreGetChannels() {
	team := model.NewId()
	o1 := model.Channel{}
	o1.TeamId = team
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = team
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	o3 := model.Channel{}
	o3.TeamId = team
	o3.DisplayName = "Channel3"
	o3.Name = "zz" + model.NewId() + "b"
	o3.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = m1.UserId
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	m4 := model.ChannelMember{}
	m4.ChannelId = o3.Id
	m4.UserId = m1.UserId
	m4.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m4)
	s.Require().Nil(err)

	list, nErr := s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, false, 0)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 3)
	s.Require().Equal(o1.Id, (*list)[0].Id, "missing channel")
	s.Require().Equal(o2.Id, (*list)[1].Id, "missing channel")
	s.Require().Equal(o3.Id, (*list)[2].Id, "missing channel")

	ids, err := s.Store().Channel().GetAllChannelMembersForUser(m1.UserId, false, false)
	s.Require().Nil(err)
	_, ok := ids[o1.Id]
	s.Require().True(ok, "missing channel")

	ids2, err := s.Store().Channel().GetAllChannelMembersForUser(m1.UserId, true, false)
	s.Require().Nil(err)
	_, ok = ids2[o1.Id]
	s.Require().True(ok, "missing channel")

	ids3, err := s.Store().Channel().GetAllChannelMembersForUser(m1.UserId, true, false)
	s.Require().Nil(err)
	_, ok = ids3[o1.Id]
	s.Require().True(ok, "missing channel")

	ids4, err := s.Store().Channel().GetAllChannelMembersForUser(m1.UserId, true, true)
	s.Require().Nil(err)
	_, ok = ids4[o1.Id]
	s.Require().True(ok, "missing channel")

	nErr = s.Store().Channel().Delete(o2.Id, 10)
	s.Require().NoError(nErr)

	nErr = s.Store().Channel().Delete(o3.Id, 20)
	s.Require().NoError(nErr)

	// should return 1
	list, nErr = s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, false, 0)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 1)
	s.Require().Equal(o1.Id, (*list)[0].Id, "missing channel")

	// Should return all
	list, nErr = s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, true, 0)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 3)
	s.Require().Equal(o1.Id, (*list)[0].Id, "missing channel")
	s.Require().Equal(o2.Id, (*list)[1].Id, "missing channel")
	s.Require().Equal(o3.Id, (*list)[2].Id, "missing channel")

	// Should still return all
	list, nErr = s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, true, 10)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 3)
	s.Require().Equal(o1.Id, (*list)[0].Id, "missing channel")
	s.Require().Equal(o2.Id, (*list)[1].Id, "missing channel")
	s.Require().Equal(o3.Id, (*list)[2].Id, "missing channel")

	// Should return 2
	list, nErr = s.Store().Channel().GetChannels(o1.TeamId, m1.UserId, true, 20)
	s.Require().Nil(nErr)
	s.Require().Len(*list, 2)
	s.Require().Equal(o1.Id, (*list)[0].Id, "missing channel")
	s.Require().Equal(o3.Id, (*list)[1].Id, "missing channel")

	s.Require().True(
		s.Store().Channel().IsUserInChannelUseCache(m1.UserId, o1.Id),
		"missing channel")
	s.Require().True(
		s.Store().Channel().IsUserInChannelUseCache(m1.UserId, o2.Id),
		"missing channel")

	s.Require().False(
		s.Store().Channel().IsUserInChannelUseCache(m1.UserId, "blahblah"),
		"missing channel")

	s.Require().False(
		s.Store().Channel().IsUserInChannelUseCache("blahblah", "blahblah"),
		"missing channel")

	s.Store().Channel().InvalidateAllChannelMembersForUser(m1.UserId)
}

func (s *ChannelStoreTestSuite) TestStoreGetAllChannels() {
	s.cleanupChannels()

	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	t2 := model.Team{}
	t2.DisplayName = "Name2"
	t2.Name = "zz" + model.NewId()
	t2.Email = MakeEmail()
	t2.Type = model.TEAM_OPEN
	_, err = s.Store().Team().Save(&t2)
	s.Require().Nil(err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1" + model.NewId()
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	group := &model.Group{
		Name:        model.NewString(model.NewId()),
		DisplayName: model.NewId(),
		Source:      model.GroupSourceLdap,
		RemoteId:    model.NewId(),
	}
	_, err = s.Store().Group().Create(group)
	s.Require().Nil(err)

	_, err = s.Store().Group().CreateGroupSyncable(model.NewGroupChannel(group.Id, c1.Id, true))
	s.Require().Nil(err)

	c2 := model.Channel{}
	c2.TeamId = t1.Id
	c2.DisplayName = "Channel2" + model.NewId()
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&c2, -1)
	s.Require().Nil(nErr)
	c2.DeleteAt = model.GetMillis()
	c2.UpdateAt = c2.DeleteAt
	nErr = s.Store().Channel().Delete(c2.Id, c2.DeleteAt)
	s.Require().Nil(nErr, "channel should have been deleted")

	c3 := model.Channel{}
	c3.TeamId = t2.Id
	c3.DisplayName = "Channel3" + model.NewId()
	c3.Name = "zz" + model.NewId() + "b"
	c3.Type = model.CHANNEL_PRIVATE
	_, nErr = s.Store().Channel().Save(&c3, -1)
	s.Require().Nil(nErr)

	u1 := model.User{Id: model.NewId()}
	u2 := model.User{Id: model.NewId()}
	_, nErr = s.Store().Channel().CreateDirectChannel(&u1, &u2)
	s.Require().Nil(nErr)

	userIds := []string{model.NewId(), model.NewId(), model.NewId()}

	c5 := model.Channel{}
	c5.Name = model.GetGroupNameFromUserIds(userIds)
	c5.DisplayName = "GroupChannel" + model.NewId()
	c5.Name = "zz" + model.NewId() + "b"
	c5.Type = model.CHANNEL_GROUP
	_, nErr = s.Store().Channel().Save(&c5, -1)
	s.Require().Nil(nErr)

	list, nErr := s.Store().Channel().GetAllChannels(0, 10, store.ChannelSearchOpts{})
	s.Require().Nil(nErr)
	s.Assert().Len(*list, 2)
	s.Assert().Equal(c1.Id, (*list)[0].Id)
	s.Assert().Equal("Name", (*list)[0].TeamDisplayName)
	s.Assert().Equal(c3.Id, (*list)[1].Id)
	s.Assert().Equal("Name2", (*list)[1].TeamDisplayName)

	count1, nErr := s.Store().Channel().GetAllChannelsCount(store.ChannelSearchOpts{})
	s.Require().Nil(nErr)

	list, nErr = s.Store().Channel().GetAllChannels(0, 10, store.ChannelSearchOpts{IncludeDeleted: true})
	s.Require().Nil(nErr)
	s.Assert().Len(*list, 3)
	s.Assert().Equal(c1.Id, (*list)[0].Id)
	s.Assert().Equal("Name", (*list)[0].TeamDisplayName)
	s.Assert().Equal(c2.Id, (*list)[1].Id)
	s.Assert().Equal(c3.Id, (*list)[2].Id)

	count2, nErr := s.Store().Channel().GetAllChannelsCount(store.ChannelSearchOpts{IncludeDeleted: true})
	s.Require().Nil(nErr)
	s.Require().True(func() bool {
		return count2 > count1
	}())

	list, nErr = s.Store().Channel().GetAllChannels(0, 1, store.ChannelSearchOpts{IncludeDeleted: true})
	s.Require().Nil(nErr)
	s.Assert().Len(*list, 1)
	s.Assert().Equal(c1.Id, (*list)[0].Id)
	s.Assert().Equal("Name", (*list)[0].TeamDisplayName)

	// Not associated to group
	list, nErr = s.Store().Channel().GetAllChannels(0, 10, store.ChannelSearchOpts{NotAssociatedToGroup: group.Id})
	s.Require().Nil(nErr)
	s.Assert().Len(*list, 1)

	// Exclude channel names
	list, nErr = s.Store().Channel().GetAllChannels(0, 10, store.ChannelSearchOpts{ExcludeChannelNames: []string{c1.Name}})
	s.Require().Nil(nErr)
	s.Assert().Len(*list, 1)

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreGetMoreChannels() {
	teamId := model.NewId()
	otherTeamId := model.NewId()
	userId := model.NewId()
	otherUserId1 := model.NewId()
	otherUserId2 := model.NewId()

	// o1 is a channel on the team to which the user (and the other user 1) belongs
	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	_, err := s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	_, err = s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      otherUserId1,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	// o2 is a channel on the other team to which the user belongs
	o2 := model.Channel{
		TeamId:      otherTeamId,
		DisplayName: "Channel2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	_, err = s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o2.Id,
		UserId:      otherUserId2,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	// o3 is a channel on the team to which the user does not belong, and thus should show up
	// in "more channels"
	o3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	// o4 is a private channel on the team to which the user does not belong
	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelB",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	// o5 is another private channel on the team to which the user does belong
	o5 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelC",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)

	_, err = s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o5.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	s.T().Run("only o3 listed in more channels", func(t *testing.T) {
		list, channelErr := s.Store().Channel().GetMoreChannels(teamId, userId, 0, 100)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o3}, list)
	})

	// o6 is another channel on the team to which the user does not belong, and would thus
	// start showing up in "more channels".
	o6 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelD",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o6, -1)
	s.Require().Nil(nErr)

	// o7 is another channel on the team to which the user does not belong, but is deleted,
	// and thus would not start showing up in "more channels"
	o7 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelD",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o7, -1)
	s.Require().Nil(nErr)

	nErr = s.Store().Channel().Delete(o7.Id, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")

	s.T().Run("both o3 and o6 listed in more channels", func(t *testing.T) {
		list, err := s.Store().Channel().GetMoreChannels(teamId, userId, 0, 100)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o3, &o6}, list)
	})

	s.T().Run("only o3 listed in more channels with offset 0, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetMoreChannels(teamId, userId, 0, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o3}, list)
	})

	s.T().Run("only o6 listed in more channels with offset 1, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetMoreChannels(teamId, userId, 1, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o6}, list)
	})

	s.T().Run("verify analytics for open channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_OPEN)
		s.Require().Nil(err)
		s.Require().EqualValues(4, count)
	})

	s.T().Run("verify analytics for private channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_PRIVATE)
		s.Require().Nil(err)
		s.Require().EqualValues(2, count)
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetPrivateChannelsForTeam() {
	teamId := model.NewId()

	// p1 is a private channel on the team
	p1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "PrivateChannel1Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr := s.Store().Channel().Save(&p1, -1)
	s.Require().Nil(nErr)

	// p2 is a private channel on another team
	p2 := model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "PrivateChannel1Team2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&p2, -1)
	s.Require().Nil(nErr)

	// o1 is a public channel on the team
	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel1Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	s.T().Run("only p1 initially listed in private channels", func(t *testing.T) {
		list, channelErr := s.Store().Channel().GetPrivateChannelsForTeam(teamId, 0, 100)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&p1}, list)
	})

	// p3 is another private channel on the team
	p3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "PrivateChannel2Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&p3, -1)
	s.Require().Nil(nErr)

	// p4 is another private, but deleted channel on the team
	p4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "PrivateChannel3Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&p4, -1)
	s.Require().Nil(nErr)
	err := s.Store().Channel().Delete(p4.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	s.T().Run("both p1 and p3 listed in private channels", func(t *testing.T) {
		list, err := s.Store().Channel().GetPrivateChannelsForTeam(teamId, 0, 100)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&p1, &p3}, list)
	})

	s.T().Run("only p1 listed in private channels with offset 0, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetPrivateChannelsForTeam(teamId, 0, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&p1}, list)
	})

	s.T().Run("only p3 listed in private channels with offset 1, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetPrivateChannelsForTeam(teamId, 1, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&p3}, list)
	})

	s.T().Run("verify analytics for private channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_PRIVATE)
		s.Require().Nil(err)
		s.Require().EqualValues(3, count)
	})

	s.T().Run("verify analytics for open open channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_OPEN)
		s.Require().Nil(err)
		s.Require().EqualValues(1, count)
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetPublicChannelsForTeam() {
	teamId := model.NewId()

	// o1 is a public channel on the team
	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel1Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	// o2 is a public channel on another team
	o2 := model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "OpenChannel1Team2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	// o3 is a private channel on the team
	o3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "PrivateChannel1Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	s.T().Run("only o1 initially listed in public channels", func(t *testing.T) {
		list, channelErr := s.Store().Channel().GetPublicChannelsForTeam(teamId, 0, 100)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o1}, list)
	})

	// o4 is another public channel on the team
	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel2Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	// o5 is another public, but deleted channel on the team
	o5 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel3Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)
	err := s.Store().Channel().Delete(o5.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	s.T().Run("both o1 and o4 listed in public channels", func(t *testing.T) {
		list, err := s.Store().Channel().GetPublicChannelsForTeam(teamId, 0, 100)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o1, &o4}, list)
	})

	s.T().Run("only o1 listed in public channels with offset 0, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetPublicChannelsForTeam(teamId, 0, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o1}, list)
	})

	s.T().Run("only o4 listed in public channels with offset 1, limit 1", func(t *testing.T) {
		list, err := s.Store().Channel().GetPublicChannelsForTeam(teamId, 1, 1)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o4}, list)
	})

	s.T().Run("verify analytics for open channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_OPEN)
		s.Require().Nil(err)
		s.Require().EqualValues(3, count)
	})

	s.T().Run("verify analytics for private channels", func(t *testing.T) {
		count, err := s.Store().Channel().AnalyticsTypeCount(teamId, model.CHANNEL_PRIVATE)
		s.Require().Nil(err)
		s.Require().EqualValues(1, count)
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetPublicChannelsByIdsForTeam() {
	teamId := model.NewId()

	// oc1 is a public channel on the team
	oc1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel1Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&oc1, -1)
	s.Require().Nil(nErr)

	// oc2 is a public channel on another team
	oc2 := model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "OpenChannel2TeamOther",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&oc2, -1)
	s.Require().Nil(nErr)

	// pc3 is a private channel on the team
	pc3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "PrivateChannel3Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&pc3, -1)
	s.Require().Nil(nErr)

	s.T().Run("oc1 by itself should be found as a public channel in the team", func(t *testing.T) {
		list, channelErr := s.Store().Channel().GetPublicChannelsByIdsForTeam(teamId, []string{oc1.Id})
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&oc1}, list)
	})

	s.T().Run("only oc1, among others, should be found as a public channel in the team", func(t *testing.T) {
		list, channelErr := s.Store().Channel().GetPublicChannelsByIdsForTeam(teamId, []string{oc1.Id, oc2.Id, model.NewId(), pc3.Id})
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&oc1}, list)
	})

	// oc4 is another public channel on the team
	oc4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel4Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&oc4, -1)
	s.Require().Nil(nErr)

	// oc4 is another public, but deleted channel on the team
	oc5 := model.Channel{
		TeamId:      teamId,
		DisplayName: "OpenChannel4Team1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&oc5, -1)
	s.Require().Nil(nErr)

	err := s.Store().Channel().Delete(oc5.Id, model.GetMillis())
	s.Require().Nil(err, "channel should have been deleted")

	s.T().Run("only oc1 and oc4, among others, should be found as a public channel in the team", func(t *testing.T) {
		list, err := s.Store().Channel().GetPublicChannelsByIdsForTeam(teamId, []string{oc1.Id, oc2.Id, model.NewId(), pc3.Id, oc4.Id})
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&oc1, &oc4}, list)
	})

	s.T().Run("random channel id should not be found as a public channel in the team", func(t *testing.T) {
		_, err := s.Store().Channel().GetPublicChannelsByIdsForTeam(teamId, []string{model.NewId()})
		s.Require().NotNil(err)
		var nfErr *store.ErrNotFound
		s.Require().True(errors.As(err, &nfErr))
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetChannelCounts() {
	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	m3 := model.ChannelMember{}
	m3.ChannelId = o2.Id
	m3.UserId = model.NewId()
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	counts, _ := s.Store().Channel().GetChannelCounts(o1.TeamId, m1.UserId)

	s.Require().Len(counts.Counts, 1, "wrong number of counts")
	s.Require().Len(counts.UpdateTimes, 1, "wrong number of update times")
}

func (s *ChannelStoreTestSuite) TestStoreGetMembersForUser() {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	o1 := model.Channel{}
	o1.TeamId = t1.Id
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	s.T().Run("with channels", func(t *testing.T) {
		var members *model.ChannelMembers
		members, err = s.Store().Channel().GetMembersForUser(o1.TeamId, m1.UserId)
		s.Require().Nil(err)

		s.Assert().Len(*members, 2)
	})

	s.T().Run("with channels and direct messages", func(t *testing.T) {
		user := model.User{Id: m1.UserId}
		u1 := model.User{Id: model.NewId()}
		u2 := model.User{Id: model.NewId()}
		u3 := model.User{Id: model.NewId()}
		u4 := model.User{Id: model.NewId()}
		_, nErr = s.Store().Channel().CreateDirectChannel(&u1, &user)
		s.Require().Nil(nErr)
		_, nErr = s.Store().Channel().CreateDirectChannel(&u2, &user)
		s.Require().Nil(nErr)
		// other user direct message
		_, nErr = s.Store().Channel().CreateDirectChannel(&u3, &u4)
		s.Require().Nil(nErr)

		var members *model.ChannelMembers
		members, err = s.Store().Channel().GetMembersForUser(o1.TeamId, m1.UserId)
		s.Require().Nil(err)

		s.Assert().Len(*members, 4)
	})

	s.T().Run("with channels, direct channels and group messages", func(t *testing.T) {
		userIds := []string{model.NewId(), model.NewId(), model.NewId(), m1.UserId}
		group := &model.Channel{
			Name:        model.GetGroupNameFromUserIds(userIds),
			DisplayName: "test",
			Type:        model.CHANNEL_GROUP,
		}
		var channel *model.Channel
		channel, nErr = s.Store().Channel().Save(group, 10000)
		s.Require().Nil(nErr)
		for _, userId := range userIds {
			cm := &model.ChannelMember{
				UserId:      userId,
				ChannelId:   channel.Id,
				NotifyProps: model.GetDefaultChannelNotifyProps(),
				SchemeUser:  true,
			}

			_, err = s.Store().Channel().SaveMember(cm)
			s.Require().Nil(err)
		}
		var members *model.ChannelMembers
		members, err = s.Store().Channel().GetMembersForUser(o1.TeamId, m1.UserId)
		s.Require().Nil(err)

		s.Assert().Len(*members, 5)
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetMembersForUserWithPagination() {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	o1 := model.Channel{}
	o1.TeamId = t1.Id
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = o1.TeamId
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	members, err := s.Store().Channel().GetMembersForUserWithPagination(o1.TeamId, m1.UserId, 0, 1)
	s.Require().Nil(err)
	s.Assert().Len(*members, 1)

	members, err = s.Store().Channel().GetMembersForUserWithPagination(o1.TeamId, m1.UserId, 1, 1)
	s.Require().Nil(err)
	s.Assert().Len(*members, 1)
}

func (s *ChannelStoreTestSuite) TestCountPostsAfter() {
	s.T().Run("should count all posts with or without the given user ID", func(t *testing.T) {
		userId1 := model.NewId()
		userId2 := model.NewId()

		channelId := model.NewId()

		p1, err := s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1000,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1001,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId2,
			ChannelId: channelId,
			CreateAt:  1002,
		})
		s.Require().Nil(err)

		count, err := s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt-1, "")
		s.Require().Nil(err)
		s.Assert().Equal(3, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt, "")
		s.Require().Nil(err)
		s.Assert().Equal(2, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt-1, userId1)
		s.Require().Nil(err)
		s.Assert().Equal(2, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt, userId1)
		s.Require().Nil(err)
		s.Assert().Equal(1, count)
	})

	s.T().Run("should not count deleted posts", func(t *testing.T) {
		userId1 := model.NewId()

		channelId := model.NewId()

		p1, err := s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1000,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1001,
			DeleteAt:  1001,
		})
		s.Require().Nil(err)

		count, err := s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt-1, "")
		s.Require().Nil(err)
		s.Assert().Equal(1, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt, "")
		s.Require().Nil(err)
		s.Assert().Equal(0, count)
	})

	s.T().Run("should count system/bot messages, but not join/leave messages", func(t *testing.T) {
		userId1 := model.NewId()

		channelId := model.NewId()

		p1, err := s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1000,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1001,
			Type:      model.POST_JOIN_CHANNEL,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1002,
			Type:      model.POST_REMOVE_FROM_CHANNEL,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1003,
			Type:      model.POST_LEAVE_TEAM,
		})
		s.Require().Nil(err)

		p5, err := s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1004,
			Type:      model.POST_HEADER_CHANGE,
		})
		s.Require().Nil(err)

		_, err = s.Store().Post().Save(&model.Post{
			UserId:    userId1,
			ChannelId: channelId,
			CreateAt:  1005,
			Type:      "custom_nps_survey",
		})
		s.Require().Nil(err)

		count, err := s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt-1, "")
		s.Require().Nil(err)
		s.Assert().Equal(3, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p1.CreateAt, "")
		s.Require().Nil(err)
		s.Assert().Equal(2, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p5.CreateAt-1, "")
		s.Require().Nil(err)
		s.Assert().Equal(2, count)

		count, err = s.Store().Channel().CountPostsAfter(channelId, p5.CreateAt, "")
		s.Require().Nil(err)
		s.Assert().Equal(1, count)
	})
}

func (s *ChannelStoreTestSuite) TestStoreUpdateLastViewedAt() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	o1.TotalMsgCount = 25
	o1.LastPostAt = 12345
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel1"
	o2.Name = "zz" + model.NewId() + "c"
	o2.Type = model.CHANNEL_OPEN
	o2.TotalMsgCount = 26
	o2.LastPostAt = 123456
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m2 := model.ChannelMember{}
	m2.ChannelId = o2.Id
	m2.UserId = m1.UserId
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	var times map[string]int64
	times, err = s.Store().Channel().UpdateLastViewedAt([]string{m1.ChannelId}, m1.UserId, false)
	s.Require().Nil(err, "failed to update ", err)
	s.Require().Equal(o1.LastPostAt, times[o1.Id], "last viewed at time incorrect")

	times, err = s.Store().Channel().UpdateLastViewedAt([]string{m1.ChannelId, m2.ChannelId}, m1.UserId, false)
	s.Require().Nil(err, "failed to update ", err)
	s.Require().Equal(o2.LastPostAt, times[o2.Id], "last viewed at time incorrect")

	rm1, err := s.Store().Channel().GetMember(m1.ChannelId, m1.UserId)
	s.Assert().Nil(err)
	s.Assert().Equal(o1.LastPostAt, rm1.LastViewedAt)
	s.Assert().Equal(o1.LastPostAt, rm1.LastUpdateAt)
	s.Assert().Equal(o1.TotalMsgCount, rm1.MsgCount)

	rm2, err := s.Store().Channel().GetMember(m2.ChannelId, m2.UserId)
	s.Assert().Nil(err)
	s.Assert().Equal(o2.LastPostAt, rm2.LastViewedAt)
	s.Assert().Equal(o2.LastPostAt, rm2.LastUpdateAt)
	s.Assert().Equal(o2.TotalMsgCount, rm2.MsgCount)

	_, err = s.Store().Channel().UpdateLastViewedAt([]string{m1.ChannelId}, "missing id", false)
	s.Require().Nil(err, "failed to update")
}

func (s *ChannelStoreTestSuite) TestStoreIncrementMentionCount() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "Channel1"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	o1.TotalMsgCount = 25
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = model.NewId()
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	err = s.Store().Channel().IncrementMentionCount(m1.ChannelId, m1.UserId, false)
	s.Require().Nil(err, "failed to update")

	err = s.Store().Channel().IncrementMentionCount(m1.ChannelId, "missing id", false)
	s.Require().Nil(err, "failed to update")

	err = s.Store().Channel().IncrementMentionCount("missing id", m1.UserId, false)
	s.Require().Nil(err, "failed to update")

	err = s.Store().Channel().IncrementMentionCount("missing id", "missing id", false)
	s.Require().Nil(err, "failed to update")
}

func (s *ChannelStoreTestSuite) TestUpdateChannelMember() {
	userId := model.NewId()

	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	m1 := &model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err := s.Store().Channel().SaveMember(m1)
	s.Require().Nil(err)

	m1.NotifyProps["test"] = "sometext"
	_, err = s.Store().Channel().UpdateMember(m1)
	s.Require().Nil(err, err)

	m1.UserId = ""
	_, err = s.Store().Channel().UpdateMember(m1)
	s.Require().NotNil(err, "bad user id - should fail")
}

func (s *ChannelStoreTestSuite) TestGetMember() {
	userId := model.NewId()

	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	c2 := &model.Channel{
		TeamId:      c1.TeamId,
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(c2, -1)
	s.Require().Nil(nErr)

	m1 := &model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err := s.Store().Channel().SaveMember(m1)
	s.Require().Nil(err)

	m2 := &model.ChannelMember{
		ChannelId:   c2.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(m2)
	s.Require().Nil(err)

	_, err = s.Store().Channel().GetMember(model.NewId(), userId)
	s.Require().NotNil(err, "should've failed to get member for non-existent channel")

	_, err = s.Store().Channel().GetMember(c1.Id, model.NewId())
	s.Require().NotNil(err, "should've failed to get member for non-existent user")

	member, err := s.Store().Channel().GetMember(c1.Id, userId)
	s.Require().Nil(err, "shouldn't have errored when getting member", err)
	s.Require().Equal(c1.Id, member.ChannelId, "should've gotten member of channel 1")
	s.Require().Equal(userId, member.UserId, "should've have gotten member for user")

	member, err = s.Store().Channel().GetMember(c2.Id, userId)
	s.Require().Nil(err, "should'nt have errored when getting member", err)
	s.Require().Equal(c2.Id, member.ChannelId, "should've gotten member of channel 2")
	s.Require().Equal(userId, member.UserId, "should've gotten member for user")

	props, err := s.Store().Channel().GetAllChannelMembersNotifyPropsForChannel(c2.Id, false)
	s.Require().Nil(err, err)
	s.Require().NotEqual(0, len(props), "should not be empty")

	props, err = s.Store().Channel().GetAllChannelMembersNotifyPropsForChannel(c2.Id, true)
	s.Require().Nil(err, err)
	s.Require().NotEqual(0, len(props), "should not be empty")

	s.Store().Channel().InvalidateCacheForChannelMembersNotifyProps(c2.Id)
}

func (s *ChannelStoreTestSuite) TestStoreGetMemberForPost() {
	ch := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}

	o1, nErr := s.Store().Channel().Save(ch, -1)
	s.Require().Nil(nErr)

	m1, err := s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	p1, nErr := s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
	})
	s.Require().Nil(nErr)

	r1, err := s.Store().Channel().GetMemberForPost(p1.Id, m1.UserId)
	s.Require().Nil(err, err)
	s.Require().Equal(m1.ToJson(), r1.ToJson(), "invalid returned channel member")

	_, err = s.Store().Channel().GetMemberForPost(p1.Id, model.NewId())
	s.Require().NotNil(err, "shouldn't have returned a member")
}

func (s *ChannelStoreTestSuite) TestGetMemberCount() {
	teamId := model.NewId()

	c1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	c2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&c2, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{
		Email:    MakeEmail(),
		DeleteAt: 0,
	}
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u1.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, nErr = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(nErr)

	count, channelErr := s.Store().Channel().GetMemberCount(c1.Id, false)
	s.Require().Nilf(channelErr, "failed to get member count: %v", channelErr)
	s.Require().EqualValuesf(1, count, "got incorrect member count %v", count)

	u2 := model.User{
		Email:    MakeEmail(),
		DeleteAt: 0,
	}
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	m2 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u2.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, nErr = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(nErr)

	count, channelErr = s.Store().Channel().GetMemberCount(c1.Id, false)
	s.Require().Nilf(channelErr, "failed to get member count: %v", channelErr)
	s.Require().EqualValuesf(2, count, "got incorrect member count %v", count)

	// make sure members of other channels aren't counted
	u3 := model.User{
		Email:    MakeEmail(),
		DeleteAt: 0,
	}
	_, err = s.Store().User().Save(&u3)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u3.Id}, -1)
	s.Require().Nil(nErr)

	m3 := model.ChannelMember{
		ChannelId:   c2.Id,
		UserId:      u3.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, nErr = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(nErr)

	count, channelErr = s.Store().Channel().GetMemberCount(c1.Id, false)
	s.Require().Nilf(channelErr, "failed to get member count: %v", channelErr)
	s.Require().EqualValuesf(2, count, "got incorrect member count %v", count)

	// make sure inactive users aren't counted
	u4 := &model.User{
		Email:    MakeEmail(),
		DeleteAt: 10000,
	}
	_, err = s.Store().User().Save(u4)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u4.Id}, -1)
	s.Require().Nil(nErr)

	m4 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u4.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, nErr = s.Store().Channel().SaveMember(&m4)
	s.Require().Nil(nErr)

	count, nErr = s.Store().Channel().GetMemberCount(c1.Id, false)
	s.Require().Nilf(nErr, "failed to get member count: %v", nErr)
	s.Require().EqualValuesf(2, count, "got incorrect member count %v", count)
}

func (s *ChannelStoreTestSuite) TestGetMemberCountsByGroup() {
	var memberCounts []*model.ChannelMemberCountByGroup
	teamId := model.NewId()
	g1 := &model.Group{
		Name:        model.NewString(model.NewId()),
		DisplayName: model.NewId(),
		Source:      model.GroupSourceLdap,
		RemoteId:    model.NewId(),
	}
	_, err := s.Store().Group().Create(g1)
	s.Require().Nil(err)

	c1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{
		Timezone: timezones.DefaultUserTimezone(),
		Email:    MakeEmail(),
		DeleteAt: 0,
	}
	_, nErr = s.Store().User().Save(u1)
	s.Require().Nil(nErr)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{
		ChannelId:   c1.Id,
		UserId:      u1.Id,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, nErr = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(nErr)

	s.T().Run("empty slice for channel with no groups", func(t *testing.T) {
		memberCounts, nErr = s.Store().Channel().GetMemberCountsByGroup(c1.Id, false)
		expectedMemberCounts := []*model.ChannelMemberCountByGroup{}
		s.Require().Nil(nErr)
		s.Require().Equal(expectedMemberCounts, memberCounts)
	})

	_, err = s.Store().Group().UpsertMember(g1.Id, u1.Id)
	s.Require().Nil(err)

	s.T().Run("returns memberCountsByGroup without timezones", func(t *testing.T) {
		memberCounts, nErr = s.Store().Channel().GetMemberCountsByGroup(c1.Id, false)
		expectedMemberCounts := []*model.ChannelMemberCountByGroup{
			{
				GroupId:                     g1.Id,
				ChannelMemberCount:          1,
				ChannelMemberTimezonesCount: 0,
			},
		}
		s.Require().Nil(nErr)
		s.Require().Equal(expectedMemberCounts, memberCounts)
	})

	s.T().Run("returns memberCountsByGroup with timezones when no timezones set", func(t *testing.T) {
		memberCounts, nErr = s.Store().Channel().GetMemberCountsByGroup(c1.Id, true)
		expectedMemberCounts := []*model.ChannelMemberCountByGroup{
			{
				GroupId:                     g1.Id,
				ChannelMemberCount:          1,
				ChannelMemberTimezonesCount: 0,
			},
		}
		s.Require().Nil(nErr)
		s.Require().Equal(expectedMemberCounts, memberCounts)
	})

	g2 := &model.Group{
		Name:        model.NewString(model.NewId()),
		DisplayName: model.NewId(),
		Source:      model.GroupSourceLdap,
		RemoteId:    model.NewId(),
	}
	_, err = s.Store().Group().Create(g2)
	s.Require().Nil(err)

	// create 5 different users with 2 different timezones for group 2
	for i := 1; i <= 5; i++ {
		timeZone := timezones.DefaultUserTimezone()
		if i == 1 {
			timeZone["manualTimezone"] = "EDT"
			timeZone["useAutomaticTimezone"] = "false"
		}

		u := &model.User{
			Timezone: timeZone,
			Email:    MakeEmail(),
			DeleteAt: 0,
		}
		_, nErr = s.Store().User().Save(u)
		s.Require().Nil(nErr)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u.Id}, -1)
		s.Require().Nil(nErr)

		m := model.ChannelMember{
			ChannelId:   c1.Id,
			UserId:      u.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}
		_, nErr = s.Store().Channel().SaveMember(&m)
		s.Require().Nil(nErr)

		_, err = s.Store().Group().UpsertMember(g2.Id, u.Id)
		s.Require().Nil(err)
	}

	g3 := &model.Group{
		Name:        model.NewString(model.NewId()),
		DisplayName: model.NewId(),
		Source:      model.GroupSourceLdap,
		RemoteId:    model.NewId(),
	}

	_, err = s.Store().Group().Create(g3)
	s.Require().Nil(err)

	// create 10 different users with 3 different timezones for group 3
	for i := 1; i <= 10; i++ {
		timeZone := timezones.DefaultUserTimezone()

		if i == 1 || i == 2 {
			timeZone["manualTimezone"] = "EDT"
			timeZone["useAutomaticTimezone"] = "false"
		} else if i == 3 || i == 4 {
			timeZone["manualTimezone"] = "PST"
			timeZone["useAutomaticTimezone"] = "false"
		} else if i == 5 || i == 6 {
			timeZone["autoTimezone"] = "PST"
			timeZone["useAutomaticTimezone"] = "true"
		} else {
			// Give every user with auto timezone set to true a random manual timezone to ensure that manual timezone is not looked at if auto is set
			timeZone["useAutomaticTimezone"] = "true"
			timeZone["manualTimezone"] = "PST" + utils.RandomName(utils.Range{Begin: 5, End: 5}, utils.ALPHANUMERIC)
		}

		u := &model.User{
			Timezone: timeZone,
			Email:    MakeEmail(),
			DeleteAt: 0,
		}
		_, nErr = s.Store().User().Save(u)
		s.Require().Nil(nErr)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u.Id}, -1)
		s.Require().Nil(nErr)

		m := model.ChannelMember{
			ChannelId:   c1.Id,
			UserId:      u.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		}
		_, nErr = s.Store().Channel().SaveMember(&m)
		s.Require().Nil(nErr)

		_, err = s.Store().Group().UpsertMember(g3.Id, u.Id)
		s.Require().Nil(err)
	}

	s.T().Run("returns memberCountsByGroup for multiple groups with lots of users without timezones", func(t *testing.T) {
		memberCounts, nErr = s.Store().Channel().GetMemberCountsByGroup(c1.Id, false)
		expectedMemberCounts := []*model.ChannelMemberCountByGroup{
			{
				GroupId:                     g1.Id,
				ChannelMemberCount:          1,
				ChannelMemberTimezonesCount: 0,
			},
			{
				GroupId:                     g2.Id,
				ChannelMemberCount:          5,
				ChannelMemberTimezonesCount: 0,
			},
			{
				GroupId:                     g3.Id,
				ChannelMemberCount:          10,
				ChannelMemberTimezonesCount: 0,
			},
		}
		s.Require().Nil(nErr)
		s.Require().ElementsMatch(expectedMemberCounts, memberCounts)
	})

	s.T().Run("returns memberCountsByGroup for multiple groups with lots of users with timezones", func(t *testing.T) {
		memberCounts, nErr = s.Store().Channel().GetMemberCountsByGroup(c1.Id, true)
		expectedMemberCounts := []*model.ChannelMemberCountByGroup{
			{
				GroupId:                     g1.Id,
				ChannelMemberCount:          1,
				ChannelMemberTimezonesCount: 0,
			},
			{
				GroupId:                     g2.Id,
				ChannelMemberCount:          5,
				ChannelMemberTimezonesCount: 1,
			},
			{
				GroupId:                     g3.Id,
				ChannelMemberCount:          10,
				ChannelMemberTimezonesCount: 3,
			},
		}
		s.Require().Nil(nErr)
		s.Require().ElementsMatch(expectedMemberCounts, memberCounts)
	})
}

func (s *ChannelStoreTestSuite) TestGetGuestCount() {
	teamId := model.NewId()

	c1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	c2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&c2, -1)
	s.Require().Nil(nErr)

	s.T().Run("Regular member doesn't count", func(t *testing.T) {
		u1 := &model.User{
			Email:    MakeEmail(),
			DeleteAt: 0,
			Roles:    model.SYSTEM_USER_ROLE_ID,
		}
		_, err := s.Store().User().Save(u1)
		s.Require().Nil(err)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u1.Id}, -1)
		s.Require().Nil(nErr)

		m1 := model.ChannelMember{
			ChannelId:   c1.Id,
			UserId:      u1.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: false,
		}
		_, nErr = s.Store().Channel().SaveMember(&m1)
		s.Require().Nil(nErr)

		count, channelErr := s.Store().Channel().GetGuestCount(c1.Id, false)
		s.Require().Nil(channelErr)
		s.Require().Equal(int64(0), count)
	})

	s.T().Run("Guest member does count", func(t *testing.T) {
		u2 := model.User{
			Email:    MakeEmail(),
			DeleteAt: 0,
			Roles:    model.SYSTEM_GUEST_ROLE_ID,
		}
		_, err := s.Store().User().Save(&u2)
		s.Require().Nil(err)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u2.Id}, -1)
		s.Require().Nil(nErr)

		m2 := model.ChannelMember{
			ChannelId:   c1.Id,
			UserId:      u2.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: true,
		}
		_, nErr = s.Store().Channel().SaveMember(&m2)
		s.Require().Nil(nErr)

		count, channelErr := s.Store().Channel().GetGuestCount(c1.Id, false)
		s.Require().Nil(channelErr)
		s.Require().Equal(int64(1), count)
	})

	s.T().Run("make sure members of other channels aren't counted", func(t *testing.T) {
		u3 := model.User{
			Email:    MakeEmail(),
			DeleteAt: 0,
			Roles:    model.SYSTEM_GUEST_ROLE_ID,
		}
		_, err := s.Store().User().Save(&u3)
		s.Require().Nil(err)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u3.Id}, -1)
		s.Require().Nil(nErr)

		m3 := model.ChannelMember{
			ChannelId:   c2.Id,
			UserId:      u3.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: true,
		}
		_, nErr = s.Store().Channel().SaveMember(&m3)
		s.Require().Nil(nErr)

		count, channelErr := s.Store().Channel().GetGuestCount(c1.Id, false)
		s.Require().Nil(channelErr)
		s.Require().Equal(int64(1), count)
	})

	s.T().Run("make sure inactive users aren't counted", func(t *testing.T) {
		u4 := &model.User{
			Email:    MakeEmail(),
			DeleteAt: 10000,
			Roles:    model.SYSTEM_GUEST_ROLE_ID,
		}
		_, err := s.Store().User().Save(u4)
		s.Require().Nil(err)
		_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: teamId, UserId: u4.Id}, -1)
		s.Require().Nil(nErr)

		m4 := model.ChannelMember{
			ChannelId:   c1.Id,
			UserId:      u4.Id,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
			SchemeGuest: true,
		}
		_, nErr = s.Store().Channel().SaveMember(&m4)
		s.Require().Nil(nErr)

		count, channelErr := s.Store().Channel().GetGuestCount(c1.Id, false)
		s.Require().Nil(channelErr)
		s.Require().Equal(int64(1), count)
	})
}

func (s *ChannelStoreTestSuite) TestStoreSearchMore() {
	teamId := model.NewId()
	otherTeamId := model.NewId()

	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	o2 := model.Channel{
		TeamId:      otherTeamId,
		DisplayName: "Channel2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m3 := model.ChannelMember{
		ChannelId:   o2.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	o3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelB",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	o5 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelC",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)

	o6 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Topic",
		Name:        "off-topic",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o6, -1)
	s.Require().Nil(nErr)

	o7 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Set",
		Name:        "off-set",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o7, -1)
	s.Require().Nil(nErr)

	o8 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Limit",
		Name:        "off-limit",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o8, -1)
	s.Require().Nil(nErr)

	o9 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel With Purpose",
		Purpose:     "This can now be searchable!",
		Name:        "with-purpose",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o9, -1)
	s.Require().Nil(nErr)

	o10 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA",
		Name:        "channel-a-deleted",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o10, -1)
	s.Require().Nil(nErr)

	o10.DeleteAt = model.GetMillis()
	o10.UpdateAt = o10.DeleteAt
	nErr = s.Store().Channel().Delete(o10.Id, o10.DeleteAt)
	s.Require().Nil(nErr, "channel should have been deleted")

	s.T().Run("three public channels matching 'ChannelA', but already a member of one and one deleted", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchMore(m1.UserId, teamId, "ChannelA")
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o3}, channels)
	})

	s.T().Run("one public channels, but already a member", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchMore(m1.UserId, teamId, o4.Name)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{}, channels)
	})

	s.T().Run("three matching channels, but only two public", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchMore(m1.UserId, teamId, "off-")
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o7, &o6}, channels)
	})

	s.T().Run("one channel matching 'off-topic'", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchMore(m1.UserId, teamId, "off-topic")
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o6}, channels)
	})

	s.T().Run("search purpose", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchMore(m1.UserId, teamId, "now searchable")
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o9}, channels)
	})
}

type ByChannelDisplayName model.ChannelList

func (s ByChannelDisplayName) Len() int { return len(s) }
func (s ByChannelDisplayName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s ByChannelDisplayName) Less(i, j int) bool {
	if s[i].DisplayName != s[j].DisplayName {
		return s[i].DisplayName < s[j].DisplayName
	}

	return s[i].Id < s[j].Id
}

func (s *ChannelStoreTestSuite) TestStoreSearchArchivedInTeam() {
	teamId := model.NewId()
	userId := model.NewId()

	s.T().Run("empty result", func(t *testing.T) {
		list, err := s.Store().Channel().SearchArchivedInTeam(teamId, "term", userId)
		s.Require().Nil(err)
		s.Require().NotNil(list)
		s.Require().Empty(list)
	})

	s.T().Run("error", func(t *testing.T) {
		// trigger a SQL error
		s.SqlStore().GetMaster().Exec("ALTER TABLE Channels RENAME TO Channels_renamed")
		defer s.SqlStore().GetMaster().Exec("ALTER TABLE Channels_renamed RENAME TO Channels")

		list, err := s.Store().Channel().SearchArchivedInTeam(teamId, "term", userId)
		s.Require().NotNil(err)
		s.Require().Nil(list)
	})
}

func (s *ChannelStoreTestSuite) TestStoreSearchInTeam() {
	teamId := model.NewId()
	otherTeamId := model.NewId()

	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{
		TeamId:      otherTeamId,
		DisplayName: "ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err := s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	m3 := model.ChannelMember{
		ChannelId:   o2.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	o3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA (alternate)",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel B",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	o5 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Channel C",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)

	o6 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Topic",
		Name:        "off-topic",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o6, -1)
	s.Require().Nil(nErr)

	o7 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Set",
		Name:        "off-set",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o7, -1)
	s.Require().Nil(nErr)

	o8 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Off-Limit",
		Name:        "off-limit",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o8, -1)
	s.Require().Nil(nErr)

	o9 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Town Square",
		Name:        "town-square",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o9, -1)
	s.Require().Nil(nErr)

	o10 := model.Channel{
		TeamId:      teamId,
		DisplayName: "The",
		Name:        "thename",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o10, -1)
	s.Require().Nil(nErr)

	o11 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Native Mobile Apps",
		Name:        "native-mobile-apps",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o11, -1)
	s.Require().Nil(nErr)

	o12 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelZ",
		Purpose:     "This can now be searchable!",
		Name:        "with-purpose",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o12, -1)
	s.Require().Nil(nErr)

	o13 := model.Channel{
		TeamId:      teamId,
		DisplayName: "ChannelA (deleted)",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o13, -1)
	s.Require().Nil(nErr)
	o13.DeleteAt = model.GetMillis()
	o13.UpdateAt = o13.DeleteAt
	nErr = s.Store().Channel().Delete(o13.Id, o13.DeleteAt)
	s.Require().Nil(nErr, "channel should have been deleted")

	testCases := []struct {
		Description     string
		TeamId          string
		Term            string
		IncludeDeleted  bool
		ExpectedResults *model.ChannelList
	}{
		{"ChannelA", teamId, "ChannelA", false, &model.ChannelList{&o1, &o3}},
		{"ChannelA, include deleted", teamId, "ChannelA", true, &model.ChannelList{&o1, &o3, &o13}},
		{"ChannelA, other team", otherTeamId, "ChannelA", false, &model.ChannelList{&o2}},
		{"empty string", teamId, "", false, &model.ChannelList{&o1, &o3, &o12, &o11, &o7, &o6, &o10, &o9}},
		{"no matches", teamId, "blargh", false, &model.ChannelList{}},
		{"prefix", teamId, "off-", false, &model.ChannelList{&o7, &o6}},
		{"full match with dash", teamId, "off-topic", false, &model.ChannelList{&o6}},
		{"town square", teamId, "town square", false, &model.ChannelList{&o9}},
		{"the in name", teamId, "thename", false, &model.ChannelList{&o10}},
		{"Mobile", teamId, "Mobile", false, &model.ChannelList{&o11}},
		{"search purpose", teamId, "now searchable", false, &model.ChannelList{&o12}},
		{"pipe ignored", teamId, "town square |", false, &model.ChannelList{&o9}},
	}

	for name, search := range map[string]func(teamId string, term string, includeDeleted bool) (*model.ChannelList, error){
		"AutocompleteInTeam": s.Store().Channel().AutocompleteInTeam,
		"SearchInTeam":       s.Store().Channel().SearchInTeam,
	} {
		for _, testCase := range testCases {
			s.T().Run(name+"/"+testCase.Description, func(t *testing.T) {
				channels, err := search(testCase.TeamId, testCase.Term, testCase.IncludeDeleted)
				s.Require().Nil(err)

				// AutoCompleteInTeam doesn't currently sort its output results.
				if name == "AutocompleteInTeam" {
					sort.Sort(ByChannelDisplayName(*channels))
				}

				s.Require().Equal(testCase.ExpectedResults, channels)
			})
		}
	}
}

func (s *ChannelStoreTestSuite) TestStoreSearchForUserInTeam() {
	userId := model.NewId()
	teamId := model.NewId()
	otherTeamId := model.NewId()

	// create 4 channels for the same team and one for other team
	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "test-dev-1",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "test-dev-2",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	o3 := model.Channel{
		TeamId:      teamId,
		DisplayName: "dev-3",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "dev-4",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	o5 := model.Channel{
		TeamId:      otherTeamId,
		DisplayName: "other-team-dev-5",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)

	// add the user to the first 3 channels and the other team channel
	for _, c := range []model.Channel{o1, o2, o3, o5} {
		_, err := s.Store().Channel().SaveMember(&model.ChannelMember{
			ChannelId:   c.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		s.Require().Nil(err)
	}

	searchAndCheck := func(term string, includeDeleted bool, expectedDisplayNames []string) {
		res, searchErr := s.Store().Channel().SearchForUserInTeam(userId, teamId, term, includeDeleted)
		s.Require().Nil(searchErr)
		s.Require().Len(*res, len(expectedDisplayNames))

		resultDisplayNames := []string{}
		for _, c := range *res {
			resultDisplayNames = append(resultDisplayNames, c.DisplayName)
		}
		s.Require().ElementsMatch(expectedDisplayNames, resultDisplayNames)
	}

	s.T().Run("Search for test, get channels 1 and 2", func(t *testing.T) {
		searchAndCheck("test", false, []string{o1.DisplayName, o2.DisplayName})
	})

	s.T().Run("Search for dev, get channels 1, 2 and 3", func(t *testing.T) {
		searchAndCheck("dev", false, []string{o1.DisplayName, o2.DisplayName, o3.DisplayName})
	})

	s.T().Run("After adding user to channel 4, search for dev, get channels 1, 2, 3 and 4", func(t *testing.T) {
		_, err := s.Store().Channel().SaveMember(&model.ChannelMember{
			ChannelId:   o4.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		s.Require().Nil(err)

		searchAndCheck("dev", false, []string{o1.DisplayName, o2.DisplayName, o3.DisplayName, o4.DisplayName})
	})

	s.T().Run("Mark channel 1 as deleted, search for dev, get channels 2, 3 and 4", func(t *testing.T) {
		o1.DeleteAt = model.GetMillis()
		o1.UpdateAt = o1.DeleteAt
		err := s.Store().Channel().Delete(o1.Id, o1.DeleteAt)
		s.Require().Nil(err)

		searchAndCheck("dev", false, []string{o2.DisplayName, o3.DisplayName, o4.DisplayName})
	})

	s.T().Run("With includeDeleted, search for dev, get channels 1, 2, 3 and 4", func(t *testing.T) {
		searchAndCheck("dev", true, []string{o1.DisplayName, o2.DisplayName, o3.DisplayName, o4.DisplayName})
	})
}

func (s *ChannelStoreTestSuite) TestStoreSearchAllChannels() {
	s.cleanupChannels()

	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	t2 := model.Team{}
	t2.DisplayName = "Name2"
	t2.Name = "zz" + model.NewId()
	t2.Email = MakeEmail()
	t2.Type = model.TEAM_OPEN
	_, err = s.Store().Team().Save(&t2)
	s.Require().Nil(err)

	o1 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A1 ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{
		TeamId:      t2.Id,
		DisplayName: "A2 ChannelA",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{
		ChannelId:   o1.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	m3 := model.ChannelMember{
		ChannelId:   o2.Id,
		UserId:      model.NewId(),
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	}
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	o3 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A3 ChannelA (alternate)",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	o4 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A4 ChannelB",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	o5 := model.Channel{
		TeamId:           t1.Id,
		DisplayName:      "A5 ChannelC",
		Name:             "zz" + model.NewId() + "b",
		Type:             model.CHANNEL_PRIVATE,
		GroupConstrained: model.NewBool(true),
	}
	_, nErr = s.Store().Channel().Save(&o5, -1)
	s.Require().Nil(nErr)

	o6 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A6 Off-Topic",
		Name:        "off-topic",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o6, -1)
	s.Require().Nil(nErr)

	o7 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A7 Off-Set",
		Name:        "off-set",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o7, -1)
	s.Require().Nil(nErr)

	group := &model.Group{
		Name:        model.NewString(model.NewId()),
		DisplayName: model.NewId(),
		Source:      model.GroupSourceLdap,
		RemoteId:    model.NewId(),
	}
	_, err = s.Store().Group().Create(group)
	s.Require().Nil(err)

	_, err = s.Store().Group().CreateGroupSyncable(model.NewGroupChannel(group.Id, o7.Id, true))
	s.Require().Nil(err)

	o8 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A8 Off-Limit",
		Name:        "off-limit",
		Type:        model.CHANNEL_PRIVATE,
	}
	_, nErr = s.Store().Channel().Save(&o8, -1)
	s.Require().Nil(nErr)

	o9 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "A9 Town Square",
		Name:        "town-square",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o9, -1)
	s.Require().Nil(nErr)

	o10 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "B10 That",
		Name:        "that",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o10, -1)
	s.Require().Nil(nErr)

	o11 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "B11 Native Mobile Apps",
		Name:        "native-mobile-apps",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o11, -1)
	s.Require().Nil(nErr)

	o12 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "B12 ChannelZ",
		Purpose:     "This can now be searchable!",
		Name:        "with-purpose",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o12, -1)
	s.Require().Nil(nErr)

	o13 := model.Channel{
		TeamId:      t1.Id,
		DisplayName: "B13 ChannelA (deleted)",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o13, -1)
	s.Require().Nil(nErr)

	o13.DeleteAt = model.GetMillis()
	o13.UpdateAt = o13.DeleteAt
	nErr = s.Store().Channel().Delete(o13.Id, o13.DeleteAt)
	s.Require().Nil(nErr, "channel should have been deleted")

	o14 := model.Channel{
		TeamId:      t2.Id,
		DisplayName: "B14 FOOBARDISPLAYNAME",
		Name:        "whatever",
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o14, -1)
	s.Require().Nil(nErr)
	testCases := []struct {
		Description     string
		Term            string
		Opts            store.ChannelSearchOpts
		ExpectedResults *model.ChannelList
		TotalCount      int
	}{
		{"Search FooBar by display name", "bardisplay", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o14}, 1},
		{"Search FooBar by display name2", "foobar", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o14}, 1},
		{"Search FooBar by display name3", "displayname", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o14}, 1},
		{"Search FooBar by name", "what", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o14}, 1},
		{"Search FooBar by name2", "ever", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o14}, 1},
		{"ChannelA", "ChannelA", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o1, &o2, &o3}, 0},
		{"ChannelA, include deleted", "ChannelA", store.ChannelSearchOpts{IncludeDeleted: true}, &model.ChannelList{&o1, &o2, &o3, &o13}, 0},
		{"empty string", "", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o1, &o2, &o3, &o4, &o5, &o6, &o7, &o8, &o9, &o10, &o11, &o12, &o14}, 0},
		{"no matches", "blargh", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{}, 0},
		{"prefix", "off-", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o6, &o7, &o8}, 0},
		{"full match with dash", "off-topic", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o6}, 0},
		{"town square", "town square", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o9}, 0},
		{"that in name", "that", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o10}, 0},
		{"Mobile", "Mobile", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o11}, 0},
		{"search purpose", "now searchable", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o12}, 0},
		{"pipe ignored", "town square |", store.ChannelSearchOpts{IncludeDeleted: false}, &model.ChannelList{&o9}, 0},
		{"exclude defaults search 'off'", "off-", store.ChannelSearchOpts{IncludeDeleted: false, ExcludeChannelNames: []string{"off-topic"}}, &model.ChannelList{&o7, &o8}, 0},
		{"exclude defaults search 'town'", "town", store.ChannelSearchOpts{IncludeDeleted: false, ExcludeChannelNames: []string{"town-square"}}, &model.ChannelList{}, 0},
		{"exclude by group association", "off-", store.ChannelSearchOpts{IncludeDeleted: false, NotAssociatedToGroup: group.Id}, &model.ChannelList{&o6, &o8}, 0},
		{"paginate includes count", "off-", store.ChannelSearchOpts{IncludeDeleted: false, PerPage: model.NewInt(100)}, &model.ChannelList{&o6, &o7, &o8}, 3},
		{"paginate, page 2 correct entries and count", "off-", store.ChannelSearchOpts{IncludeDeleted: false, PerPage: model.NewInt(2), Page: model.NewInt(1)}, &model.ChannelList{&o8}, 3},
		{"Filter private", "", store.ChannelSearchOpts{IncludeDeleted: false, Private: true}, &model.ChannelList{&o4, &o5, &o8}, 3},
		{"Filter public", "", store.ChannelSearchOpts{IncludeDeleted: false, Public: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o6, &o7}, 10},
		{"Filter public and private", "", store.ChannelSearchOpts{IncludeDeleted: false, Public: true, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o4, &o5}, 13},
		{"Filter public and private and include deleted", "", store.ChannelSearchOpts{IncludeDeleted: true, Public: true, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o4, &o5}, 14},
		{"Filter group constrained", "", store.ChannelSearchOpts{IncludeDeleted: false, GroupConstrained: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o5}, 1},
		{"Filter exclude group constrained and include deleted", "", store.ChannelSearchOpts{IncludeDeleted: true, ExcludeGroupConstrained: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o4, &o6}, 13},
		{"Filter private and exclude group constrained", "", store.ChannelSearchOpts{IncludeDeleted: false, ExcludeGroupConstrained: true, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o4, &o8}, 2},
		{"Filter team 2", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t2.Id}, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o2, &o14}, 2},
		{"Filter team 2, private", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t2.Id}, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{}, 0},
		{"Filter team 1 and team 2, private", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t1.Id, t2.Id}, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o4, &o5, &o8}, 3},
		{"Filter team 1 and team 2, public and private", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t1.Id, t2.Id}, Public: true, Private: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o4, &o5}, 13},
		{"Filter team 1 and team 2, public and private and group constrained", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t1.Id, t2.Id}, Public: true, Private: true, GroupConstrained: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o5}, 1},
		{"Filter team 1 and team 2, public and private and exclude group constrained", "", store.ChannelSearchOpts{IncludeDeleted: false, TeamIds: []string{t1.Id, t2.Id}, Public: true, Private: true, ExcludeGroupConstrained: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o1, &o2, &o3, &o4, &o6}, 12},
		{"Filter deleted returns only deleted channels", "", store.ChannelSearchOpts{Deleted: true, Page: model.NewInt(0), PerPage: model.NewInt(5)}, &model.ChannelList{&o13}, 1},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.Description, func(t *testing.T) {
			channels, count, err := s.Store().Channel().SearchAllChannels(testCase.Term, testCase.Opts)
			s.Require().Nil(err)
			s.Require().Equal(len(*testCase.ExpectedResults), len(*channels))
			for i, expected := range *testCase.ExpectedResults {
				s.Require().Equal(expected.Id, (*channels)[i].Id)
			}
			if testCase.Opts.Page != nil || testCase.Opts.PerPage != nil {
				s.Require().Equal(int64(testCase.TotalCount), count)
			}
		})
	}
}

func (s *ChannelStoreTestSuite) TestStoreGetMembersByIds() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "ChannelA"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	m1 := &model.ChannelMember{ChannelId: o1.Id, UserId: model.NewId(), NotifyProps: model.GetDefaultChannelNotifyProps()}
	_, err := s.Store().Channel().SaveMember(m1)
	s.Require().Nil(err)

	var members *model.ChannelMembers
	members, nErr = s.Store().Channel().GetMembersByIds(m1.ChannelId, []string{m1.UserId})
	s.Require().Nil(nErr, nErr)
	rm1 := (*members)[0]

	s.Require().Equal(m1.ChannelId, rm1.ChannelId, "bad team id")
	s.Require().Equal(m1.UserId, rm1.UserId, "bad user id")

	m2 := &model.ChannelMember{ChannelId: o1.Id, UserId: model.NewId(), NotifyProps: model.GetDefaultChannelNotifyProps()}
	_, err = s.Store().Channel().SaveMember(m2)
	s.Require().Nil(err)

	members, nErr = s.Store().Channel().GetMembersByIds(m1.ChannelId, []string{m1.UserId, m2.UserId, model.NewId()})
	s.Require().Nil(nErr, nErr)
	s.Require().Len(*members, 2, "return wrong number of results")

	_, nErr = s.Store().Channel().GetMembersByIds(m1.ChannelId, []string{})
	s.Require().NotNil(nErr, "empty user ids - should have failed")
}

func (s *ChannelStoreTestSuite) TestStoreGetMembersByChannelIds() {
	userId := model.NewId()

	// Create a couple channels and add the user to them
	channel1, err := s.Store().Channel().Save(&model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}, -1)
	s.Require().Nil(err)

	channel2, err := s.Store().Channel().Save(&model.Channel{
		TeamId:      model.NewId(),
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}, -1)
	s.Require().Nil(err)

	_, err = s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   channel1.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	_, err = s.Store().Channel().SaveMember(&model.ChannelMember{
		ChannelId:   channel2.Id,
		UserId:      userId,
		NotifyProps: model.GetDefaultChannelNotifyProps(),
	})
	s.Require().Nil(err)

	s.T().Run("should return the user's members for the given channels", func(t *testing.T) {
		result, nErr := s.Store().Channel().GetMembersByChannelIds([]string{channel1.Id, channel2.Id}, userId)
		s.Require().Nil(nErr)
		s.Assert().Len(*result, 2)

		s.Assert().Equal(userId, (*result)[0].UserId)
		s.Assert().True((*result)[0].ChannelId == channel1.Id || (*result)[1].ChannelId == channel1.Id)
		s.Assert().Equal(userId, (*result)[1].UserId)
		s.Assert().True((*result)[0].ChannelId == channel2.Id || (*result)[1].ChannelId == channel2.Id)
	})

	s.T().Run("should not error or return anything for invalid channel IDs", func(t *testing.T) {
		result, nErr := s.Store().Channel().GetMembersByChannelIds([]string{model.NewId(), model.NewId()}, userId)
		s.Require().Nil(nErr)
		s.Assert().Len(*result, 0)
	})

	s.T().Run("should not error or return anything for invalid user IDs", func(t *testing.T) {
		result, nErr := s.Store().Channel().GetMembersByChannelIds([]string{channel1.Id, channel2.Id}, model.NewId())
		s.Require().Nil(nErr)
		s.Assert().Len(*result, 0)
	})
}

func (s *ChannelStoreTestSuite) TestStoreSearchGroupChannels() {
	// Users
	u1 := &model.User{}
	u1.Username = "user.one"
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)

	u2 := &model.User{}
	u2.Username = "user.two"
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)

	u3 := &model.User{}
	u3.Username = "user.three"
	u3.Email = MakeEmail()
	u3.Nickname = model.NewId()
	_, err = s.Store().User().Save(u3)
	s.Require().Nil(err)

	u4 := &model.User{}
	u4.Username = "user.four"
	u4.Email = MakeEmail()
	u4.Nickname = model.NewId()
	_, err = s.Store().User().Save(u4)
	s.Require().Nil(err)

	// Group channels
	userIds := []string{u1.Id, u2.Id, u3.Id}
	gc1 := model.Channel{}
	gc1.Name = model.GetGroupNameFromUserIds(userIds)
	gc1.DisplayName = "GroupChannel" + model.NewId()
	gc1.Type = model.CHANNEL_GROUP
	_, nErr := s.Store().Channel().Save(&gc1, -1)
	s.Require().Nil(nErr)

	for _, userId := range userIds {
		_, nErr = s.Store().Channel().SaveMember(&model.ChannelMember{
			ChannelId:   gc1.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		s.Require().Nil(nErr)
	}

	userIds = []string{u1.Id, u4.Id}
	gc2 := model.Channel{}
	gc2.Name = model.GetGroupNameFromUserIds(userIds)
	gc2.DisplayName = "GroupChannel" + model.NewId()
	gc2.Type = model.CHANNEL_GROUP
	_, nErr = s.Store().Channel().Save(&gc2, -1)
	s.Require().Nil(nErr)

	for _, userId := range userIds {
		_, err := s.Store().Channel().SaveMember(&model.ChannelMember{
			ChannelId:   gc2.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		s.Require().Nil(err)
	}

	userIds = []string{u1.Id, u2.Id, u3.Id, u4.Id}
	gc3 := model.Channel{}
	gc3.Name = model.GetGroupNameFromUserIds(userIds)
	gc3.DisplayName = "GroupChannel" + model.NewId()
	gc3.Type = model.CHANNEL_GROUP
	_, nErr = s.Store().Channel().Save(&gc3, -1)
	s.Require().Nil(nErr)

	for _, userId := range userIds {
		_, err := s.Store().Channel().SaveMember(&model.ChannelMember{
			ChannelId:   gc3.Id,
			UserId:      userId,
			NotifyProps: model.GetDefaultChannelNotifyProps(),
		})
		s.Require().Nil(err)
	}

	defer func() {
		for _, gc := range []model.Channel{gc1, gc2, gc3} {
			s.Store().Channel().PermanentDeleteMembersByChannel(gc3.Id)
			s.Store().Channel().PermanentDelete(gc.Id)
		}
	}()

	testCases := []struct {
		Name           string
		UserId         string
		Term           string
		ExpectedResult []string
	}{
		{
			Name:           "Get all group channels for user1",
			UserId:         u1.Id,
			Term:           "",
			ExpectedResult: []string{gc1.Id, gc2.Id, gc3.Id},
		},
		{
			Name:           "Get group channels for user1 and term 'three'",
			UserId:         u1.Id,
			Term:           "three",
			ExpectedResult: []string{gc1.Id, gc3.Id},
		},
		{
			Name:           "Get group channels for user1 and term 'four two'",
			UserId:         u1.Id,
			Term:           "four two",
			ExpectedResult: []string{gc3.Id},
		},
		{
			Name:           "Get all group channels for user2",
			UserId:         u2.Id,
			Term:           "",
			ExpectedResult: []string{gc1.Id, gc3.Id},
		},
		{
			Name:           "Get group channels for user2 and term 'four'",
			UserId:         u2.Id,
			Term:           "four",
			ExpectedResult: []string{gc3.Id},
		},
		{
			Name:           "Get all group channels for user4",
			UserId:         u4.Id,
			Term:           "",
			ExpectedResult: []string{gc2.Id, gc3.Id},
		},
		{
			Name:           "Get group channels for user4 and term 'one five'",
			UserId:         u4.Id,
			Term:           "one five",
			ExpectedResult: []string{},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			result, err := s.Store().Channel().SearchGroupChannels(tc.UserId, tc.Term)
			s.Require().Nil(err)

			resultIds := []string{}
			for _, gc := range *result {
				resultIds = append(resultIds, gc.Id)
			}

			s.Require().ElementsMatch(tc.ExpectedResult, resultIds)
		})
	}
}

func (s *ChannelStoreTestSuite) TestStoreAnalyticsDeletedTypeCount() {
	o1 := model.Channel{}
	o1.TeamId = model.NewId()
	o1.DisplayName = "ChannelA"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	o2 := model.Channel{}
	o2.TeamId = model.NewId()
	o2.DisplayName = "Channel2"
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	p3 := model.Channel{}
	p3.TeamId = model.NewId()
	p3.DisplayName = "Channel3"
	p3.Name = "zz" + model.NewId() + "b"
	p3.Type = model.CHANNEL_PRIVATE
	_, nErr = s.Store().Channel().Save(&p3, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)

	d4, nErr := s.Store().Channel().CreateDirectChannel(u1, u2)
	s.Require().Nil(nErr)
	defer func() {
		s.Store().Channel().PermanentDeleteMembersByChannel(d4.Id)
		s.Store().Channel().PermanentDelete(d4.Id)
	}()

	var openStartCount int64
	openStartCount, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "O")
	s.Require().Nil(nErr, nErr)

	var privateStartCount int64
	privateStartCount, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "P")
	s.Require().Nil(nErr, nErr)

	var directStartCount int64
	directStartCount, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "D")
	s.Require().Nil(nErr, nErr)

	nErr = s.Store().Channel().Delete(o1.Id, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")
	nErr = s.Store().Channel().Delete(o2.Id, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")
	nErr = s.Store().Channel().Delete(p3.Id, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")
	nErr = s.Store().Channel().Delete(d4.Id, model.GetMillis())
	s.Require().Nil(nErr, "channel should have been deleted")

	var count int64

	count, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "O")
	s.Require().Nil(err, nErr)
	s.Assert().Equal(openStartCount+2, count, "Wrong open channel deleted count.")

	count, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "P")
	s.Require().Nil(nErr, nErr)
	s.Assert().Equal(privateStartCount+1, count, "Wrong private channel deleted count.")

	count, nErr = s.Store().Channel().AnalyticsDeletedTypeCount("", "D")
	s.Require().Nil(nErr, nErr)
	s.Assert().Equal(directStartCount+1, count, "Wrong direct channel deleted count.")
}

func (s *ChannelStoreTestSuite) TestStoreGetPinnedPosts() {
	ch1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}

	o1, nErr := s.Store().Channel().Save(ch1, -1)
	s.Require().Nil(nErr)

	p1, err := s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
		IsPinned:  true,
	})
	s.Require().Nil(err)

	pl, errGet := s.Store().Channel().GetPinnedPosts(o1.Id)
	s.Require().Nil(errGet, errGet)
	s.Require().NotNil(pl.Posts[p1.Id], "didn't return relevant pinned posts")

	ch2 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}

	o2, nErr := s.Store().Channel().Save(ch2, -1)
	s.Require().Nil(nErr)

	_, err = s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o2.Id,
		Message:   "test",
	})
	s.Require().Nil(err)

	pl, errGet = s.Store().Channel().GetPinnedPosts(o2.Id)
	s.Require().Nil(errGet, errGet)
	s.Require().Empty(pl.Posts, "wasn't supposed to return posts")

	s.T().Run("with correct ReplyCount", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
			IsPinned:  true,
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post2, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
			IsPinned:  true,
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post3, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post1.Id,
			RootId:    post1.Id,
			Message:   "message",
			IsPinned:  true,
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		posts, err := s.Store().Channel().GetPinnedPosts(channelId)
		s.Require().Nil(err)
		s.Require().Len(posts.Posts, 3)
		s.Require().Equal(posts.Posts[post1.Id].ReplyCount, int64(1))
		s.Require().Equal(posts.Posts[post2.Id].ReplyCount, int64(0))
		s.Require().Equal(posts.Posts[post3.Id].ReplyCount, int64(1))
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetPinnedPostCount() {
	ch1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}

	o1, nErr := s.Store().Channel().Save(ch1, -1)
	s.Require().Nil(nErr)

	_, err := s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
		IsPinned:  true,
	})
	s.Require().Nil(err)

	_, err = s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o1.Id,
		Message:   "test",
		IsPinned:  true,
	})
	s.Require().Nil(err)

	count, errGet := s.Store().Channel().GetPinnedPostCount(o1.Id, true)
	s.Require().Nil(errGet, errGet)
	s.Require().EqualValues(2, count, "didn't return right count")

	ch2 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        "zz" + model.NewId() + "b",
		Type:        model.CHANNEL_OPEN,
	}

	o2, nErr := s.Store().Channel().Save(ch2, -1)
	s.Require().Nil(nErr)

	_, err = s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o2.Id,
		Message:   "test",
	})
	s.Require().Nil(err)

	_, err = s.Store().Post().Save(&model.Post{
		UserId:    model.NewId(),
		ChannelId: o2.Id,
		Message:   "test",
	})
	s.Require().Nil(err)

	count, errGet = s.Store().Channel().GetPinnedPostCount(o2.Id, true)
	s.Require().Nil(errGet, errGet)
	s.Require().EqualValues(0, count, "should return 0")
}

func (s *ChannelStoreTestSuite) TestStoreMaxChannelsPerTeam() {
	channel := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Channel",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(channel, 0)
	s.Assert().NotNil(nErr)
	var ltErr *store.ErrLimitExceeded
	s.Assert().True(errors.As(nErr, &ltErr))

	channel.Id = ""
	_, nErr = s.Store().Channel().Save(channel, 1)
	s.Assert().Nil(nErr)
}

func (s *ChannelStoreTestSuite) TestStoreGetChannelsByScheme() {
	// Create some schemes.
	s1 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}

	s2 := &model.Scheme{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}

	s1, err := s.Store().Scheme().Save(s1)
	s.Require().Nil(err)
	s2, err = s.Store().Scheme().Save(s2)
	s.Require().Nil(err)

	// Create and save some teams.
	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &s1.Id,
	}

	c2 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &s1.Id,
	}

	c3 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}

	_, _ = s.Store().Channel().Save(c1, 100)
	_, _ = s.Store().Channel().Save(c2, 100)
	_, _ = s.Store().Channel().Save(c3, 100)

	// Get the channels by a valid Scheme ID.
	d1, err := s.Store().Channel().GetChannelsByScheme(s1.Id, 0, 100)
	s.Assert().Nil(err)
	s.Assert().Len(d1, 2)

	// Get the channels by a valid Scheme ID where there aren't any matching Channel.
	d2, err := s.Store().Channel().GetChannelsByScheme(s2.Id, 0, 100)
	s.Assert().Nil(err)
	s.Assert().Empty(d2)

	// Get the channels by an invalid Scheme ID.
	d3, err := s.Store().Channel().GetChannelsByScheme(model.NewId(), 0, 100)
	s.Assert().Nil(err)
	s.Assert().Empty(d3)
}

func (s *ChannelStoreTestSuite) TestStoreMigrateChannelMembers() {
	s1 := model.NewId()
	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &s1,
	}
	c1, _ = s.Store().Channel().Save(c1, 100)

	cm1 := &model.ChannelMember{
		ChannelId:     c1.Id,
		UserId:        model.NewId(),
		ExplicitRoles: "channel_admin channel_user",
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
	}
	cm2 := &model.ChannelMember{
		ChannelId:     c1.Id,
		UserId:        model.NewId(),
		ExplicitRoles: "channel_user",
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
	}
	cm3 := &model.ChannelMember{
		ChannelId:     c1.Id,
		UserId:        model.NewId(),
		ExplicitRoles: "something_else",
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
	}

	cm1, _ = s.Store().Channel().SaveMember(cm1)
	cm2, _ = s.Store().Channel().SaveMember(cm2)
	cm3, _ = s.Store().Channel().SaveMember(cm3)

	lastDoneChannelId := strings.Repeat("0", 26)
	lastDoneUserId := strings.Repeat("0", 26)

	for {
		data, err := s.Store().Channel().MigrateChannelMembers(lastDoneChannelId, lastDoneUserId)
		if s.Assert().Nil(err) {
			if data == nil {
				break
			}
			lastDoneChannelId = data["ChannelId"]
			lastDoneUserId = data["UserId"]
		}
	}

	s.Store().Channel().ClearCaches()

	cm1b, err := s.Store().Channel().GetMember(cm1.ChannelId, cm1.UserId)
	s.Assert().Nil(err)
	s.Assert().Equal("", cm1b.ExplicitRoles)
	s.Assert().False(cm1b.SchemeGuest)
	s.Assert().True(cm1b.SchemeUser)
	s.Assert().True(cm1b.SchemeAdmin)

	cm2b, err := s.Store().Channel().GetMember(cm2.ChannelId, cm2.UserId)
	s.Assert().Nil(err)
	s.Assert().Equal("", cm2b.ExplicitRoles)
	s.Assert().False(cm1b.SchemeGuest)
	s.Assert().True(cm2b.SchemeUser)
	s.Assert().False(cm2b.SchemeAdmin)

	cm3b, err := s.Store().Channel().GetMember(cm3.ChannelId, cm3.UserId)
	s.Assert().Nil(err)
	s.Assert().Equal("something_else", cm3b.ExplicitRoles)
	s.Assert().False(cm1b.SchemeGuest)
	s.Assert().False(cm3b.SchemeUser)
	s.Assert().False(cm3b.SchemeAdmin)
}

func (s *ChannelStoreTestSuite) TestResetAllChannelSchemes() {
	s1 := &model.Scheme{
		Name:        model.NewId(),
		DisplayName: model.NewId(),
		Description: model.NewId(),
		Scope:       model.SCHEME_SCOPE_CHANNEL,
	}
	s1, err := s.Store().Scheme().Save(s1)
	s.Require().Nil(err)

	c1 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &s1.Id,
	}

	c2 := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
		SchemeId:    &s1.Id,
	}

	c1, _ = s.Store().Channel().Save(c1, 100)
	c2, _ = s.Store().Channel().Save(c2, 100)

	s.Assert().Equal(s1.Id, *c1.SchemeId)
	s.Assert().Equal(s1.Id, *c2.SchemeId)

	err = s.Store().Channel().ResetAllChannelSchemes()
	s.Assert().Nil(err)

	c1, _ = s.Store().Channel().Get(c1.Id, true)
	c2, _ = s.Store().Channel().Get(c2.Id, true)

	s.Assert().Equal("", *c1.SchemeId)
	s.Assert().Equal("", *c2.SchemeId)
}

func (s *ChannelStoreTestSuite) TestStoreClearAllCustomRoleAssignments() {
	c := &model.Channel{
		TeamId:      model.NewId(),
		DisplayName: "Name",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}

	c, _ = s.Store().Channel().Save(c, 100)

	m1 := &model.ChannelMember{
		ChannelId:     c.Id,
		UserId:        model.NewId(),
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
		ExplicitRoles: "system_user_access_token channel_user channel_admin",
	}
	m2 := &model.ChannelMember{
		ChannelId:     c.Id,
		UserId:        model.NewId(),
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
		ExplicitRoles: "channel_user custom_role channel_admin another_custom_role",
	}
	m3 := &model.ChannelMember{
		ChannelId:     c.Id,
		UserId:        model.NewId(),
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
		ExplicitRoles: "channel_user",
	}
	m4 := &model.ChannelMember{
		ChannelId:     c.Id,
		UserId:        model.NewId(),
		NotifyProps:   model.GetDefaultChannelNotifyProps(),
		ExplicitRoles: "custom_only",
	}

	_, err := s.Store().Channel().SaveMember(m1)
	s.Require().Nil(err)
	_, err = s.Store().Channel().SaveMember(m2)
	s.Require().Nil(err)
	_, err = s.Store().Channel().SaveMember(m3)
	s.Require().Nil(err)
	_, err = s.Store().Channel().SaveMember(m4)
	s.Require().Nil(err)

	s.Require().Nil(s.Store().Channel().ClearAllCustomRoleAssignments())

	member, err := s.Store().Channel().GetMember(m1.ChannelId, m1.UserId)
	s.Require().Nil(err)
	s.Assert().Equal(m1.ExplicitRoles, member.Roles)

	member, err = s.Store().Channel().GetMember(m2.ChannelId, m2.UserId)
	s.Require().Nil(err)
	s.Assert().Equal("channel_user channel_admin", member.Roles)

	member, err = s.Store().Channel().GetMember(m3.ChannelId, m3.UserId)
	s.Require().Nil(err)
	s.Assert().Equal(m3.ExplicitRoles, member.Roles)

	member, err = s.Store().Channel().GetMember(m4.ChannelId, m4.UserId)
	s.Require().Nil(err)
	s.Assert().Equal("", member.Roles)
}

// testMaterializedPublicChannels tests edge cases involving the triggers and stored procedures
// that materialize the PublicChannels table.
func (s *ChannelStoreTestSuite) TestMaterializedPublicChannels() {
	teamId := model.NewId()

	// o1 is a public channel on the team
	o1 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Open Channel",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr := s.Store().Channel().Save(&o1, -1)
	s.Require().Nil(nErr)

	// o2 is another public channel on the team
	o2 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Open Channel 2",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}
	_, nErr = s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	s.T().Run("o1 and o2 initially listed in public channels", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o1, &o2}, channels)
	})

	o1.DeleteAt = model.GetMillis()
	o1.UpdateAt = o1.DeleteAt

	e := s.Store().Channel().Delete(o1.Id, o1.DeleteAt)
	s.Require().Nil(e, "channel should have been deleted")

	s.T().Run("o1 still listed in public channels when marked as deleted", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o1, &o2}, channels)
	})

	s.Store().Channel().PermanentDelete(o1.Id)

	s.T().Run("o1 no longer listed in public channels when permanently deleted", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o2}, channels)
	})

	o2.Type = model.CHANNEL_PRIVATE
	_, appErr := s.Store().Channel().Update(&o2)
	s.Require().Nil(appErr)

	s.T().Run("o2 no longer listed since now private", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{}, channels)
	})

	o2.Type = model.CHANNEL_OPEN
	_, appErr = s.Store().Channel().Update(&o2)
	s.Require().Nil(appErr)

	s.T().Run("o2 listed once again since now public", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o2}, channels)
	})

	// o3 is a public channel on the team that already existed in the PublicChannels table.
	o3 := model.Channel{
		Id:          model.NewId(),
		TeamId:      teamId,
		DisplayName: "Open Channel 3",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}

	_, execerr := s.SqlStore().GetMaster().ExecNoTimeout(`
		INSERT INTO
		    PublicChannels(Id, DeleteAt, TeamId, DisplayName, Name, Header, Purpose)
		VALUES
		    (:Id, :DeleteAt, :TeamId, :DisplayName, :Name, :Header, :Purpose);
	`, map[string]interface{}{
		"Id":          o3.Id,
		"DeleteAt":    o3.DeleteAt,
		"TeamId":      o3.TeamId,
		"DisplayName": o3.DisplayName,
		"Name":        o3.Name,
		"Header":      o3.Header,
		"Purpose":     o3.Purpose,
	})
	s.Require().Nil(execerr)

	o3.DisplayName = "Open Channel 3 - Modified"

	_, execerr = s.SqlStore().GetMaster().ExecNoTimeout(`
		INSERT INTO
		    Channels(Id, CreateAt, UpdateAt, DeleteAt, TeamId, Type, DisplayName, Name, Header, Purpose, LastPostAt, TotalMsgCount, ExtraUpdateAt, CreatorId)
		VALUES
		    (:Id, :CreateAt, :UpdateAt, :DeleteAt, :TeamId, :Type, :DisplayName, :Name, :Header, :Purpose, :LastPostAt, :TotalMsgCount, :ExtraUpdateAt, :CreatorId);
	`, map[string]interface{}{
		"Id":            o3.Id,
		"CreateAt":      o3.CreateAt,
		"UpdateAt":      o3.UpdateAt,
		"DeleteAt":      o3.DeleteAt,
		"TeamId":        o3.TeamId,
		"Type":          o3.Type,
		"DisplayName":   o3.DisplayName,
		"Name":          o3.Name,
		"Header":        o3.Header,
		"Purpose":       o3.Purpose,
		"LastPostAt":    o3.LastPostAt,
		"TotalMsgCount": o3.TotalMsgCount,
		"ExtraUpdateAt": o3.ExtraUpdateAt,
		"CreatorId":     o3.CreatorId,
	})
	s.Require().Nil(execerr)

	s.T().Run("verify o3 INSERT converted to UPDATE", func(t *testing.T) {
		channels, channelErr := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(channelErr)
		s.Require().Equal(&model.ChannelList{&o2, &o3}, channels)
	})

	// o4 is a public channel on the team that existed in the Channels table but was omitted from the PublicChannels table.
	o4 := model.Channel{
		TeamId:      teamId,
		DisplayName: "Open Channel 4",
		Name:        model.NewId(),
		Type:        model.CHANNEL_OPEN,
	}

	_, nErr = s.Store().Channel().Save(&o4, -1)
	s.Require().Nil(nErr)

	_, execerr = s.SqlStore().GetMaster().ExecNoTimeout(`
		DELETE FROM
		    PublicChannels
		WHERE
		    Id = :Id
	`, map[string]interface{}{
		"Id": o4.Id,
	})
	s.Require().Nil(execerr)

	o4.DisplayName += " - Modified"
	_, appErr = s.Store().Channel().Update(&o4)
	s.Require().Nil(appErr)

	s.T().Run("verify o4 UPDATE converted to INSERT", func(t *testing.T) {
		channels, err := s.Store().Channel().SearchInTeam(teamId, "", true)
		s.Require().Nil(err)
		s.Require().Equal(&model.ChannelList{&o2, &o3, &o4}, channels)
	})
}

func (s *ChannelStoreTestSuite) TestStoreGetAllChannelsForExportAfter() {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	d1, err := s.Store().Channel().GetAllChannelsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(err)

	found := false
	for _, c := range d1 {
		if c.Id == c1.Id {
			found = true
			s.Assert().Equal(t1.Id, c.TeamId)
			s.Assert().Nil(c.SchemeId)
			s.Assert().Equal(t1.Name, c.TeamName)
		}
	}
	s.Assert().True(found)
}

func (s *ChannelStoreTestSuite) TestStoreGetChannelMembersForExport() {
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	c2 := model.Channel{}
	c2.TeamId = model.NewId()
	c2.DisplayName = "Channel2"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(&c2, -1)
	s.Require().Nil(nErr)

	u1 := model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u1)
	s.Require().Nil(err)

	m1 := model.ChannelMember{}
	m1.ChannelId = c1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = c2.Id
	m2.UserId = u1.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	d1, err := s.Store().Channel().GetChannelMembersForExport(u1.Id, t1.Id)
	s.Assert().Nil(err)

	s.Assert().Len(d1, 1)

	cmfe1 := d1[0]
	s.Assert().Equal(c1.Name, cmfe1.ChannelName)
	s.Assert().Equal(c1.Id, cmfe1.ChannelId)
	s.Assert().Equal(u1.Id, cmfe1.UserId)
}

func (s *ChannelStoreTestSuite) TestStoreRemoveAllDeactivatedMembers() {
	// Set up all the objects needed in the store.
	t1 := model.Team{}
	t1.DisplayName = "Name"
	t1.Name = "zz" + model.NewId()
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	_, err := s.Store().Team().Save(&t1)
	s.Require().Nil(err)

	c1 := model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&c1, -1)
	s.Require().Nil(nErr)

	u1 := model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u1)
	s.Require().Nil(err)

	u2 := model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u2)
	s.Require().Nil(err)

	u3 := model.User{}
	u3.Email = MakeEmail()
	u3.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u3)
	s.Require().Nil(err)

	m1 := model.ChannelMember{}
	m1.ChannelId = c1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m1)
	s.Require().Nil(err)

	m2 := model.ChannelMember{}
	m2.ChannelId = c1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m2)
	s.Require().Nil(err)

	m3 := model.ChannelMember{}
	m3.ChannelId = c1.Id
	m3.UserId = u3.Id
	m3.NotifyProps = model.GetDefaultChannelNotifyProps()
	_, err = s.Store().Channel().SaveMember(&m3)
	s.Require().Nil(err)

	// Get all the channel members. Check there are 3.
	d1, err := s.Store().Channel().GetMembers(c1.Id, 0, 1000)
	s.Assert().Nil(err)
	s.Assert().Len(*d1, 3)

	// Deactivate users 1 & 2.
	u1.DeleteAt = model.GetMillis()
	u2.DeleteAt = model.GetMillis()
	_, err = s.Store().User().Update(&u1, true)
	s.Require().Nil(err)
	_, err = s.Store().User().Update(&u2, true)
	s.Require().Nil(err)

	// Remove all deactivated users from the channel.
	s.Assert().Nil(s.Store().Channel().RemoveAllDeactivatedMembers(c1.Id))

	// Get all the channel members. Check there is now only 1: m3.
	d2, err := s.Store().Channel().GetMembers(c1.Id, 0, 1000)
	s.Assert().Nil(err)
	s.Assert().Len(*d2, 1)
	s.Assert().Equal(u3.Id, (*d2)[0].UserId)

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreExportAllDirectChannels() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name" + model.NewId()
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	userIds := []string{model.NewId(), model.NewId(), model.NewId()}

	o2 := model.Channel{}
	o2.Name = model.GetGroupNameFromUserIds(userIds)
	o2.DisplayName = "GroupChannel" + model.NewId()
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_GROUP
	_, nErr := s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)

	d1, nErr := s.Store().Channel().GetAllDirectChannelsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(nErr)

	s.Assert().Len(d1, 2)
	s.Assert().ElementsMatch([]string{o1.DisplayName, o2.DisplayName}, []string{d1[0].DisplayName, d1[1].DisplayName})

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreExportAllDirectChannelsExcludePrivateAndPublic() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "The Direct Channel" + model.NewId()
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	o2 := model.Channel{}
	o2.TeamId = teamId
	o2.DisplayName = "Channel2" + model.NewId()
	o2.Name = "zz" + model.NewId() + "b"
	o2.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(&o2, -1)
	s.Require().Nil(nErr)

	o3 := model.Channel{}
	o3.TeamId = teamId
	o3.DisplayName = "Channel3" + model.NewId()
	o3.Name = "zz" + model.NewId() + "b"
	o3.Type = model.CHANNEL_PRIVATE
	_, nErr = s.Store().Channel().Save(&o3, -1)
	s.Require().Nil(nErr)

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)

	d1, nErr := s.Store().Channel().GetAllDirectChannelsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(nErr)
	s.Assert().Len(d1, 1)
	s.Assert().Equal(o1.DisplayName, d1[0].DisplayName)

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreExportAllDirectChannelsDeletedChannel() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Different Name" + model.NewId()
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.Email = MakeEmail()
	u2.Nickname = model.NewId()
	_, err = s.Store().User().Save(u2)
	s.Require().Nil(err)
	_, nErr = s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u2.Id}, -1)
	s.Require().Nil(nErr)

	m1 := model.ChannelMember{}
	m1.ChannelId = o1.Id
	m1.UserId = u1.Id
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := model.ChannelMember{}
	m2.ChannelId = o1.Id
	m2.UserId = u2.Id
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)

	o1.DeleteAt = 1
	nErr = s.Store().Channel().SetDeleteAt(o1.Id, 1, 1)
	s.Require().Nil(nErr, "channel should have been deleted")

	d1, nErr := s.Store().Channel().GetAllDirectChannelsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(nErr)

	s.Assert().Equal(0, len(d1))

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *ChannelStoreTestSuite) TestStoreGetChannelsBatchForIndexing() {
	// Set up all the objects needed
	c1 := &model.Channel{}
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	_, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	time.Sleep(10 * time.Millisecond)

	c2 := &model.Channel{}
	c2.DisplayName = "Channel2"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(c2, -1)
	s.Require().Nil(nErr)

	time.Sleep(10 * time.Millisecond)
	startTime := c2.CreateAt

	c3 := &model.Channel{}
	c3.DisplayName = "Channel3"
	c3.Name = "zz" + model.NewId() + "b"
	c3.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(c3, -1)
	s.Require().Nil(nErr)

	c4 := &model.Channel{}
	c4.DisplayName = "Channel4"
	c4.Name = "zz" + model.NewId() + "b"
	c4.Type = model.CHANNEL_PRIVATE
	_, nErr = s.Store().Channel().Save(c4, -1)
	s.Require().Nil(nErr)

	c5 := &model.Channel{}
	c5.DisplayName = "Channel5"
	c5.Name = "zz" + model.NewId() + "b"
	c5.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(c5, -1)
	s.Require().Nil(nErr)

	time.Sleep(10 * time.Millisecond)

	c6 := &model.Channel{}
	c6.DisplayName = "Channel6"
	c6.Name = "zz" + model.NewId() + "b"
	c6.Type = model.CHANNEL_OPEN
	_, nErr = s.Store().Channel().Save(c6, -1)
	s.Require().Nil(nErr)

	endTime := c6.CreateAt

	// First and last channel should be outside the range
	channels, err := s.Store().Channel().GetChannelsBatchForIndexing(startTime, endTime, 1000)
	s.Assert().Nil(err)
	s.Assert().ElementsMatch([]*model.Channel{c2, c3, c5}, channels)

	// Update the endTime, last channel should be in
	endTime = model.GetMillis()
	channels, err = s.Store().Channel().GetChannelsBatchForIndexing(startTime, endTime, 1000)
	s.Assert().Nil(err)
	s.Assert().ElementsMatch([]*model.Channel{c2, c3, c5, c6}, channels)

	// Testing the limit
	channels, err = s.Store().Channel().GetChannelsBatchForIndexing(startTime, endTime, 2)
	s.Assert().Nil(err)
	s.Assert().ElementsMatch([]*model.Channel{c2, c3}, channels)
}

func (s *ChannelStoreTestSuite) TestGroupSyncedChannelCount() {
	channel1, nErr := s.Store().Channel().Save(&model.Channel{
		DisplayName:      model.NewId(),
		Name:             model.NewId(),
		Type:             model.CHANNEL_PRIVATE,
		GroupConstrained: model.NewBool(true),
	}, 999)
	s.Require().Nil(nErr)
	s.Require().True(channel1.IsGroupConstrained())
	defer s.Store().Channel().PermanentDelete(channel1.Id)

	channel2, nErr := s.Store().Channel().Save(&model.Channel{
		DisplayName: model.NewId(),
		Name:        model.NewId(),
		Type:        model.CHANNEL_PRIVATE,
	}, 999)
	s.Require().Nil(nErr)
	s.Require().False(channel2.IsGroupConstrained())
	defer s.Store().Channel().PermanentDelete(channel2.Id)

	count, appErr := s.Store().Channel().GroupSyncedChannelCount()
	s.Require().Nil(appErr)
	s.Require().GreaterOrEqual(count, int64(1))

	channel2.GroupConstrained = model.NewBool(true)
	channel2, err := s.Store().Channel().Update(channel2)
	s.Require().Nil(err)
	s.Require().True(channel2.IsGroupConstrained())

	countAfter, appErr := s.Store().Channel().GroupSyncedChannelCount()
	s.Require().Nil(appErr)
	s.Require().GreaterOrEqual(countAfter, count+1)
}
