package sqlstore

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type PostStoreTestSuite struct {
	suite.Suite
	StoreTestSuite
}

func TestPostStoreTestSuite(t *testing.T) {
	StoreTestSuiteWithSqlSupplier(t, &PostStoreTestSuite{}, func(t *testing.T, testSuite StoreTestBaseSuite) {
		suite.Run(t, testSuite)
	})
}

func generateMultiplePosts() (*model.Post, *model.Post, *model.Post, *model.Post) {
	p1 := &model.Post{}
	p1.ChannelId = model.NewId()
	p1.UserId = model.NewId()
	p1.Message = "zz" + model.NewId() + "b"

	p2 := &model.Post{}
	p2.ChannelId = model.NewId()
	p2.UserId = model.NewId()
	p2.Message = "zz" + model.NewId() + "b"

	p3 := &model.Post{}
	p3.ChannelId = model.NewId()
	p3.UserId = model.NewId()
	p3.Message = "zz" + model.NewId() + "b"

	p4 := &model.Post{}
	p4.ChannelId = model.NewId()
	p4.UserId = model.NewId()
	p4.Message = "zz" + model.NewId() + "b"
	return p1, p2, p3, p4
}

// Save correctly a new set of posts
func (s *PostStoreTestSuite) TestPostStoreSaveMultiple_SaveCorrectlyNewSetOfPosts() {
	p1, p2, p3, _ := generateMultiplePosts()

	newPosts, errIdx, err := s.Store().Post().SaveMultiple([]*model.Post{p1, p2, p3})
	s.Require().Nil(err)
	s.Require().Equal(-1, errIdx)
	for _, post := range newPosts {
		storedPost, err := s.Store().Post().GetSingle(post.Id)
		s.Nil(err)
		s.Equal(post.ChannelId, storedPost.ChannelId)
		s.Equal(post.Message, storedPost.Message)
		s.Equal(post.UserId, storedPost.UserId)
	}
}

// Save replies
func (s *PostStoreTestSuite) TestPostStoreSaveMultiple_SaveReplies() {
	o1 := model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.RootId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	o2 := model.Post{}
	o2.ChannelId = model.NewId()
	o2.UserId = model.NewId()
	o2.RootId = o1.RootId
	o2.Message = "zz" + model.NewId() + "b"

	o3 := model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.RootId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"

	o4 := model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"

	newPosts, errIdx, err := s.Store().Post().SaveMultiple([]*model.Post{&o1, &o2, &o3, &o4})
	s.Require().Nil(err, "couldn't save item")
	s.Require().Equal(-1, errIdx)
	s.Len(newPosts, 4)
	s.Equal(int64(2), newPosts[0].ReplyCount)
	s.Equal(int64(2), newPosts[1].ReplyCount)
	s.Equal(int64(1), newPosts[2].ReplyCount)
	s.Equal(int64(0), newPosts[3].ReplyCount)
}

// Try to save mixed, already saved and not saved posts
func (s *PostStoreTestSuite) TestPostStoreSaveMultiple_SaveMixed() {
	_, _, p3, p4 := generateMultiplePosts()

	newPosts, errIdx, err := s.Store().Post().SaveMultiple([]*model.Post{p3})
	s.Require().NoError(err)
	s.Require().Equal(-1, errIdx)

	newPosts, errIdx, err = s.Store().Post().SaveMultiple([]*model.Post{p4, p3})
	s.Require().Error(err)
	s.Require().Equal(1, errIdx)
	s.Require().Nil(newPosts)
	storedPost, err := s.Store().Post().GetSingle(p3.Id)
	s.NoError(err)
	s.Equal(p3.ChannelId, storedPost.ChannelId)
	s.Equal(p3.Message, storedPost.Message)
	s.Equal(p3.UserId, storedPost.UserId)

	storedPost, err = s.Store().Post().GetSingle(p4.Id)
	s.Error(err)
	s.Nil(storedPost)
}

// Update reply should update the UpdateAt of the root post
func (s *PostStoreTestSuite) TestPostStoreSaveMultiple_UpdateReplyWithUpdateAt() {
	rootPost := model.Post{}
	rootPost.ChannelId = model.NewId()
	rootPost.UserId = model.NewId()
	rootPost.Message = "zz" + model.NewId() + "b"

	replyPost := model.Post{}
	replyPost.ChannelId = rootPost.ChannelId
	replyPost.UserId = model.NewId()
	replyPost.Message = "zz" + model.NewId() + "b"
	replyPost.RootId = rootPost.Id

	_, _, err := s.Store().Post().SaveMultiple([]*model.Post{&rootPost, &replyPost})
	s.Require().Nil(err)

	rrootPost, err := s.Store().Post().GetSingle(rootPost.Id)
	s.Require().Nil(err)
	s.Equal(rrootPost.UpdateAt, rootPost.UpdateAt)

	replyPost2 := model.Post{}
	replyPost2.ChannelId = rootPost.ChannelId
	replyPost2.UserId = model.NewId()
	replyPost2.Message = "zz" + model.NewId() + "b"
	replyPost2.RootId = rootPost.Id

	replyPost3 := model.Post{}
	replyPost3.ChannelId = rootPost.ChannelId
	replyPost3.UserId = model.NewId()
	replyPost3.Message = "zz" + model.NewId() + "b"
	replyPost3.RootId = rootPost.Id

	_, _, err = s.Store().Post().SaveMultiple([]*model.Post{&replyPost2, &replyPost3})
	s.Require().Nil(err)

	rrootPost2, err := s.Store().Post().GetSingle(rootPost.Id)
	s.Require().Nil(err)
	s.Greater(rrootPost2.UpdateAt, rrootPost.UpdateAt)
}

// Create a post should update the channel LastPostAt and the total messages count by one
func (s *PostStoreTestSuite) TestPostStoreSaveMultiple_CreateWithLastPostAt() {
	channel := model.Channel{}
	channel.Name = "zz" + model.NewId() + "b"
	channel.DisplayName = "zz" + model.NewId() + "b"
	channel.Type = model.CHANNEL_OPEN

	_, err := s.Store().Channel().Save(&channel, 100)
	s.Require().Nil(err)

	post1 := model.Post{}
	post1.ChannelId = channel.Id
	post1.UserId = model.NewId()
	post1.Message = "zz" + model.NewId() + "b"

	post2 := model.Post{}
	post2.ChannelId = channel.Id
	post2.UserId = model.NewId()
	post2.Message = "zz" + model.NewId() + "b"
	post2.CreateAt = 5

	post3 := model.Post{}
	post3.ChannelId = channel.Id
	post3.UserId = model.NewId()
	post3.Message = "zz" + model.NewId() + "b"

	_, _, err = s.Store().Post().SaveMultiple([]*model.Post{&post1, &post2, &post3})
	s.Require().Nil(err)

	rchannel, err := s.Store().Channel().Get(channel.Id, false)
	s.Require().Nil(err)
	s.Greater(rchannel.LastPostAt, channel.LastPostAt)
	s.Equal(int64(3), rchannel.TotalMsgCount)
}

func (s *PostStoreTestSuite) TestPostStoreSavePost_Success() {
	o1 := model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	p, err := s.Store().Post().Save(&o1)
	s.Require().Nil(err, "couldn't save item")
	s.Equal(int64(0), p.ReplyCount)
}

func (s *PostStoreTestSuite) TestPostStoreSavePost_SaveReplies() {
	o1 := model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.RootId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	o2 := model.Post{}
	o2.ChannelId = model.NewId()
	o2.UserId = model.NewId()
	o2.RootId = o1.RootId
	o2.Message = "zz" + model.NewId() + "b"

	o3 := model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.RootId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"

	p1, err := s.Store().Post().Save(&o1)
	s.Require().Nil(err, "couldn't save item")
	s.Equal(int64(1), p1.ReplyCount)

	p2, err := s.Store().Post().Save(&o2)
	s.Require().Nil(err, "couldn't save item")
	s.Equal(int64(2), p2.ReplyCount)

	p3, err := s.Store().Post().Save(&o3)
	s.Require().Nil(err, "couldn't save item")
	s.Equal(int64(1), p3.ReplyCount)
}

// Try to save existing post
func (s *PostStoreTestSuite) TestPostStoreSavePost_SaveExistingPost() {
	o1 := model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	_, err := s.Store().Post().Save(&o1)
	s.Require().Nil(err, "couldn't save item")

	_, err = s.Store().Post().Save(&o1)
	s.Require().NotNil(err, "shouldn't be able to update from save")
}

// Update reply should update the UpdateAt of the root post
func (s *PostStoreTestSuite) TestPostStoreSavePost_UpdateReplyUpdateRoot() {
	rootPost := model.Post{}
	rootPost.ChannelId = model.NewId()
	rootPost.UserId = model.NewId()
	rootPost.Message = "zz" + model.NewId() + "b"

	_, err := s.Store().Post().Save(&rootPost)
	s.Require().NoError(err)

	time.Sleep(2 * time.Millisecond)

	replyPost := model.Post{}
	replyPost.ChannelId = rootPost.ChannelId
	replyPost.UserId = model.NewId()
	replyPost.Message = "zz" + model.NewId() + "b"
	replyPost.RootId = rootPost.Id

	// We need to sleep here to be sure the post is not created during the same millisecond
	time.Sleep(time.Millisecond)
	_, err = s.Store().Post().Save(&replyPost)
	s.Require().NoError(err)

	rrootPost, err := s.Store().Post().GetSingle(rootPost.Id)
	s.Require().NoError(err)
	s.Greater(rrootPost.UpdateAt, rootPost.UpdateAt)
}

// Create a post should update the channel LastPostAt and the total messages count by one
func (s *PostStoreTestSuite) TestPostStoreSavePost_CreatePostUpdateChannelLastPostAt() {
	channel := model.Channel{}
	channel.Name = "zz" + model.NewId() + "b"
	channel.DisplayName = "zz" + model.NewId() + "b"
	channel.Type = model.CHANNEL_OPEN

	_, err := s.Store().Channel().Save(&channel, 100)
	s.Require().NoError(err)

	post := model.Post{}
	post.ChannelId = channel.Id
	post.UserId = model.NewId()
	post.Message = "zz" + model.NewId() + "b"

	// We need to sleep here to be sure the post is not created during the same millisecond
	time.Sleep(time.Millisecond)
	_, err = s.Store().Post().Save(&post)
	s.Require().NoError(err)

	rchannel, err := s.Store().Channel().Get(channel.Id, false)
	s.Require().NoError(err)
	s.Greater(rchannel.LastPostAt, channel.LastPostAt)
	s.Equal(int64(1), rchannel.TotalMsgCount)

	post = model.Post{}
	post.ChannelId = channel.Id
	post.UserId = model.NewId()
	post.Message = "zz" + model.NewId() + "b"
	post.CreateAt = 5

	// We need to sleep here to be sure the post is not created during the same millisecond
	time.Sleep(time.Millisecond)
	_, err = s.Store().Post().Save(&post)
	s.Require().NoError(err)

	rchannel2, err := s.Store().Channel().Get(channel.Id, false)
	s.Require().NoError(err)
	s.Equal(rchannel.LastPostAt, rchannel2.LastPostAt)
	s.Equal(int64(2), rchannel2.TotalMsgCount)

	post = model.Post{}
	post.ChannelId = channel.Id
	post.UserId = model.NewId()
	post.Message = "zz" + model.NewId() + "b"

	// We need to sleep here to be sure the post is not created during the same millisecond
	time.Sleep(time.Millisecond)
	_, err = s.Store().Post().Save(&post)
	s.Require().NoError(err)

	rchannel3, err := s.Store().Channel().Get(channel.Id, false)
	s.Require().NoError(err)
	s.Greater(rchannel3.LastPostAt, rchannel2.LastPostAt)
	s.Equal(int64(3), rchannel3.TotalMsgCount)
}

func (s *PostStoreTestSuite) TestPostStoreGetSingle() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"

	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	post, err := s.Store().Post().GetSingle(o1.Id)
	s.Require().Nil(err)
	s.Require().Equal(post.CreateAt, o1.CreateAt, "invalid returned post")

	_, err = s.Store().Post().GetSingle("123")
	s.Require().NotNil(err, "Missing id should have failed")
}