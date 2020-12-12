package sqlstore

import (
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/masterhung0112/hk_server/utils"
	"github.com/stretchr/testify/suite"
	"sort"
	"strings"
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

func (s *PostStoreTestSuite) TestUpdate() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	r1, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro1 := r1.Posts[o1.Id]

	r2, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro2 := r2.Posts[o2.Id]

	r3, err := s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err)
	ro3 := r3.Posts[o3.Id]

	s.Require().Equal(ro1.Message, o1.Message, "Failed to save/get")

	o1a := ro1.Clone()
	o1a.Message = ro1.Message + "BBBBBBBBBB"
	_, err = s.Store().Post().Update(o1a, ro1)
	s.Require().Nil(err)

	r1, err = s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)

	ro1a := r1.Posts[o1.Id]
	s.Require().Equal(ro1a.Message, o1a.Message, "Failed to update/get")

	o2a := ro2.Clone()
	o2a.Message = ro2.Message + "DDDDDDD"
	_, err = s.Store().Post().Update(o2a, ro2)
	s.Require().Nil(err)

	r2, err = s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro2a := r2.Posts[o2.Id]

	s.Require().Equal(ro2a.Message, o2a.Message, "Failed to update/get")

	o3a := ro3.Clone()
	o3a.Message = ro3.Message + "WWWWWWW"
	_, err = s.Store().Post().Update(o3a, ro3)
	s.Require().Nil(err)

	r3, err = s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err)
	ro3a := r3.Posts[o3.Id]

	if ro3a.Message != o3a.Message {
		s.Require().Equal(ro3a.Hashtags, o3a.Hashtags, "Failed to update/get")
	}

	o4, err := s.Store().Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test"},
	})
	s.Require().Nil(err)

	r4, err := s.Store().Post().Get(o4.Id, false)
	s.Require().Nil(err)
	ro4 := r4.Posts[o4.Id]

	o4a := ro4.Clone()
	o4a.Filenames = []string{}
	o4a.FileIds = []string{model.NewId()}
	_, err = s.Store().Post().Update(o4a, ro4)
	s.Require().Nil(err)

	r4, err = s.Store().Post().Get(o4.Id, false)
	s.Require().Nil(err)

	ro4a := r4.Posts[o4.Id]
	s.Require().Empty(ro4a.Filenames, "Failed to clear Filenames")
	s.Require().Len(ro4a.FileIds, 1, "Failed to set FileIds")
}

func (s *PostStoreTestSuite) TestDelete() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	deleteByID := model.NewId()

	etag1 := s.Store().Post().GetEtag(o1.ChannelId, false)
	s.Require().Equal(0, strings.Index(etag1, model.CurrentVersion+"."), "Invalid Etag")

	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	r1, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	s.Require().Equal(r1.Posts[o1.Id].CreateAt, o1.CreateAt, "invalid returned post")

	err = s.Store().Post().Delete(o1.Id, model.GetMillis(), deleteByID)
	s.Require().Nil(err)

	posts, _ := s.Store().Post().GetPostsCreatedAt(o1.ChannelId, o1.CreateAt)
	post := posts[0]
	actual := post.GetProp(model.POST_PROPS_DELETE_BY)

	s.Assert().Equal(deleteByID, actual, "Expected (*Post).Props[model.POST_PROPS_DELETE_BY] to be %v but got %v.", deleteByID, actual)

	r3, err := s.Store().Post().Get(o1.Id, false)
	s.Require().NotNil(err, "Missing id should have failed - PostList %v", r3)

	etag2 := s.Store().Post().GetEtag(o1.ChannelId, false)
	s.Require().Equal(0, strings.Index(etag2, model.CurrentVersion+"."), "Invalid Etag")
}

func (s *PostStoreTestSuite) TestDelete1Level() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	err = s.Store().Post().Delete(o1.Id, model.GetMillis(), "")
	s.Require().Nil(err)

	_, err = s.Store().Post().Get(o1.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o2.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")
}

func (s *PostStoreTestSuite) TestDelete2Level() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o2.Id
	o3.RootId = o1.Id
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4, err = s.Store().Post().Save(o4)
	s.Require().Nil(err)

	err = s.Store().Post().Delete(o1.Id, model.GetMillis(), "")
	s.Require().Nil(err)

	_, err = s.Store().Post().Get(o1.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o2.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o3.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o4.Id, false)
	s.Require().Nil(err)
}

func (s *PostStoreTestSuite) TestPermDelete1Level() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	err2 := s.Store().Post().PermanentDeleteByUser(o2.UserId)
	s.Require().Nil(err2)

	_, err = s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err, "Deleted id shouldn't have failed")

	_, err = s.Store().Post().Get(o2.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	err = s.Store().Post().PermanentDeleteByChannel(o3.ChannelId)
	s.Require().Nil(err)

	_, err = s.Store().Post().Get(o3.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")
}

func (s *PostStoreTestSuite) TestPermDelete1Level2() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	err2 := s.Store().Post().PermanentDeleteByUser(o1.UserId)
	s.Require().Nil(err2)

	_, err = s.Store().Post().Get(o1.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o2.Id, false)
	s.Require().NotNil(err, "Deleted id should have failed")

	_, err = s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err, "Deleted id should have failed")
}

func (s *PostStoreTestSuite) TestGetWithChildren() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o2.Id
	o3.RootId = o1.Id
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	pl, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)

	s.Require().Len(pl.Posts, 3, "invalid returned post")

	dErr := s.Store().Post().Delete(o3.Id, model.GetMillis(), "")
	s.Require().Nil(dErr)

	pl, err = s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)

	s.Require().Len(pl.Posts, 2, "invalid returned post")

	dErr = s.Store().Post().Delete(o2.Id, model.GetMillis(), "")
	s.Require().Nil(dErr)

	pl, err = s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)

	s.Require().Len(pl.Posts, 1, "invalid returned post")
}

func (s *PostStoreTestSuite) TestGetPostsWithDetails() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	_, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2a := &model.Post{}
	o2a.ChannelId = o1.ChannelId
	o2a.UserId = model.NewId()
	o2a.Message = "zz" + model.NewId() + "b"
	o2a.ParentId = o1.Id
	o2a.RootId = o1.Id
	o2a, err = s.Store().Post().Save(o2a)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.ParentId = o1.Id
	o3.RootId = o1.Id
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = o1.ChannelId
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4, err = s.Store().Post().Save(o4)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o5 := &model.Post{}
	o5.ChannelId = o1.ChannelId
	o5.UserId = model.NewId()
	o5.Message = "zz" + model.NewId() + "b"
	o5.ParentId = o4.Id
	o5.RootId = o4.Id
	o5, err = s.Store().Post().Save(o5)
	s.Require().Nil(err)

	r1, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: o1.ChannelId, Page: 0, PerPage: 4}, false)
	s.Require().Nil(err)

	s.Require().Equal(r1.Order[0], o5.Id, "invalid order")
	s.Require().Equal(r1.Order[1], o4.Id, "invalid order")
	s.Require().Equal(r1.Order[2], o3.Id, "invalid order")
	s.Require().Equal(r1.Order[3], o2a.Id, "invalid order")

	//the last 4, + o1 (o2a and o3's parent) + o2 (in same thread as o2a and o3)
	s.Require().Len(r1.Posts, 6, "wrong size")

	s.Require().Equal(r1.Posts[o1.Id].Message, o1.Message, "Missing parent")

	r2, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: o1.ChannelId, Page: 0, PerPage: 4}, false)
	s.Require().Nil(err)

	s.Require().Equal(r2.Order[0], o5.Id, "invalid order")
	s.Require().Equal(r2.Order[1], o4.Id, "invalid order")
	s.Require().Equal(r2.Order[2], o3.Id, "invalid order")
	s.Require().Equal(r2.Order[3], o2a.Id, "invalid order")

	//the last 4, + o1 (o2a and o3's parent) + o2 (in same thread as o2a and o3)
	s.Require().Len(r2.Posts, 6, "wrong size")

	s.Require().Equal(r2.Posts[o1.Id].Message, o1.Message, "Missing parent")

	// Run once to fill cache
	_, err = s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: o1.ChannelId, Page: 0, PerPage: 30}, false)
	s.Require().Nil(err)

	o6 := &model.Post{}
	o6.ChannelId = o1.ChannelId
	o6.UserId = model.NewId()
	o6.Message = "zz" + model.NewId() + "b"
	_, err = s.Store().Post().Save(o6)
	s.Require().Nil(err)

	r3, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: o1.ChannelId, Page: 0, PerPage: 30}, false)
	s.Require().Nil(err)
	s.Assert().Equal(7, len(r3.Order))
}

func (s *PostStoreTestSuite) TestGetPostsBeforeAfter() {
	s.T().Run("without threads", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		var posts []*model.Post
		for i := 0; i < 10; i++ {
			post, err := s.Store().Post().Save(&model.Post{
				ChannelId: channelId,
				UserId:    userId,
				Message:   "message",
			})
			s.Require().Nil(err)

			posts = append(posts, post)

			time.Sleep(time.Millisecond)
		}

		s.T().Run("should return error if negative Page/PerPage options are passed", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: posts[0].Id, Page: 0, PerPage: -1})
			s.Assert().Nil(postList)
			s.Assert().Error(err)
			s.Assert().IsType(&store.ErrInvalidInput{}, err)

			postList, err = s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: posts[0].Id, Page: -1, PerPage: 10})
			s.Assert().Nil(postList)
			s.Assert().Error(err)
			s.Assert().IsType(&store.ErrInvalidInput{}, err)
		})

		s.T().Run("should not return anything before the first post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: posts[0].Id, Page: 0, PerPage: 10})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{}, postList.Posts)
		})

		s.T().Run("should return posts before a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: posts[5].Id, Page: 0, PerPage: 10})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{posts[4].Id, posts[3].Id, posts[2].Id, posts[1].Id, posts[0].Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				posts[0].Id: posts[0],
				posts[1].Id: posts[1],
				posts[2].Id: posts[2],
				posts[3].Id: posts[3],
				posts[4].Id: posts[4],
			}, postList.Posts)
		})

		s.T().Run("should limit posts before", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: posts[5].Id, PerPage: 2})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{posts[4].Id, posts[3].Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				posts[3].Id: posts[3],
				posts[4].Id: posts[4],
			}, postList.Posts)
		})

		s.T().Run("should not return anything after the last post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: posts[len(posts)-1].Id, PerPage: 10})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{}, postList.Posts)
		})

		s.T().Run("should return posts after a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: posts[5].Id, PerPage: 10})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{posts[9].Id, posts[8].Id, posts[7].Id, posts[6].Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				posts[6].Id: posts[6],
				posts[7].Id: posts[7],
				posts[8].Id: posts[8],
				posts[9].Id: posts[9],
			}, postList.Posts)
		})

		s.T().Run("should limit posts after", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: posts[5].Id, PerPage: 2})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{posts[7].Id, posts[6].Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				posts[6].Id: posts[6],
				posts[7].Id: posts[7],
			}, postList.Posts)
		})
	})
	s.T().Run("with threads", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		// This creates a series of posts that looks like:
		// post1
		// post2
		// post3 (in response to post1)
		// post4 (in response to post2)
		// post5
		// post6 (in response to post2)

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		post1.ReplyCount = 1
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post2, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		post2.ReplyCount = 2
		time.Sleep(time.Millisecond)

		post3, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post1.Id,
			RootId:    post1.Id,
			Message:   "message",
		})
		s.Require().Nil(err)
		post3.ReplyCount = 1
		time.Sleep(time.Millisecond)

		post4, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			RootId:    post2.Id,
			ParentId:  post2.Id,
			Message:   "message",
		})
		s.Require().Nil(err)
		post4.ReplyCount = 2
		time.Sleep(time.Millisecond)

		post5, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post6, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post2.Id,
			RootId:    post2.Id,
			Message:   "message",
		})
		post6.ReplyCount = 2
		s.Require().Nil(err)

		// Adding a post to a thread changes the UpdateAt timestamp of the parent post
		post1.UpdateAt = post3.UpdateAt
		post2.UpdateAt = post6.UpdateAt

		s.T().Run("should return each post and thread before a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: post4.Id, PerPage: 2})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{post3.Id, post2.Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				post1.Id: post1,
				post2.Id: post2,
				post3.Id: post3,
				post4.Id: post4,
				post6.Id: post6,
			}, postList.Posts)
		})

		s.T().Run("should return each post and the root of each thread after a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: post4.Id, PerPage: 2})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{post6.Id, post5.Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				post2.Id: post2,
				post4.Id: post4,
				post5.Id: post5,
				post6.Id: post6,
			}, postList.Posts)
		})
	})
	s.T().Run("with threads (skipFetchThreads)", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		// This creates a series of posts that looks like:
		// post1
		// post2
		// post3 (in response to post1)
		// post4 (in response to post2)
		// post5
		// post6 (in response to post2)

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "post1",
		})
		s.Require().Nil(err)
		post1.ReplyCount = 1
		time.Sleep(time.Millisecond)

		post2, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "post2",
		})
		s.Require().Nil(err)
		post2.ReplyCount = 2
		time.Sleep(time.Millisecond)

		post3, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post1.Id,
			RootId:    post1.Id,
			Message:   "post3",
		})
		s.Require().Nil(err)
		post3.ReplyCount = 1
		time.Sleep(time.Millisecond)

		post4, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			RootId:    post2.Id,
			ParentId:  post2.Id,
			Message:   "post4",
		})
		s.Require().Nil(err)
		post4.ReplyCount = 2
		time.Sleep(time.Millisecond)

		post5, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "post5",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post6, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			ParentId:  post2.Id,
			RootId:    post2.Id,
			Message:   "post6",
		})
		post6.ReplyCount = 2
		s.Require().Nil(err)

		// Adding a post to a thread changes the UpdateAt timestamp of the parent post
		post1.UpdateAt = post3.UpdateAt
		post2.UpdateAt = post6.UpdateAt

		s.T().Run("should return each post and thread before a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: post4.Id, PerPage: 2, SkipFetchThreads: true})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{post3.Id, post2.Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				post1.Id: post1,
				post2.Id: post2,
				post3.Id: post3,
			}, postList.Posts)
		})

		s.T().Run("should return each post and thread before a post with limit", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsBefore(model.GetPostsOptions{ChannelId: channelId, PostId: post4.Id, PerPage: 1, SkipFetchThreads: true})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{post3.Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				post1.Id: post1,
				post3.Id: post3,
			}, postList.Posts)
		})

		s.T().Run("should return each post and the root of each thread after a post", func(t *testing.T) {
			postList, err := s.Store().Post().GetPostsAfter(model.GetPostsOptions{ChannelId: channelId, PostId: post4.Id, PerPage: 2, SkipFetchThreads: true})
			s.Assert().Nil(err)

			s.Assert().Equal([]string{post6.Id, post5.Id}, postList.Order)
			s.Assert().Equal(map[string]*model.Post{
				post2.Id: post2,
				post5.Id: post5,
				post6.Id: post6,
			}, postList.Posts)
		})
	})
}

func (s *PostStoreTestSuite) TestGetPostsSince() {
	s.T().Run("should return posts created after the given time", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		_, err = s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post3, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post4, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post5, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
			RootId:    post3.Id,
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		post6, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
			RootId:    post1.Id,
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		postList, err := s.Store().Post().GetPostsSince(model.GetPostsSinceOptions{ChannelId: channelId, Time: post3.CreateAt}, false)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{
			post6.Id,
			post5.Id,
			post4.Id,
			post3.Id,
			post1.Id,
		}, postList.Order)

		s.Assert().Len(postList.Posts, 5)
		s.Assert().NotNil(postList.Posts[post1.Id], "should return the parent post")
		s.Assert().NotNil(postList.Posts[post3.Id])
		s.Assert().NotNil(postList.Posts[post4.Id])
		s.Assert().NotNil(postList.Posts[post5.Id])
		s.Assert().NotNil(postList.Posts[post6.Id])
	})

	s.T().Run("should return empty list when nothing has changed", func(t *testing.T) {
		channelId := model.NewId()
		userId := model.NewId()

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		postList, err := s.Store().Post().GetPostsSince(model.GetPostsSinceOptions{ChannelId: channelId, Time: post1.CreateAt}, false)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{}, postList.Order)
		s.Assert().Empty(postList.Posts)
	})

	s.T().Run("should not cache a timestamp of 0 when nothing has changed", func(t *testing.T) {
		s.Store().Post().ClearCaches()

		channelId := model.NewId()
		userId := model.NewId()

		post1, err := s.Store().Post().Save(&model.Post{
			ChannelId: channelId,
			UserId:    userId,
			Message:   "message",
		})
		s.Require().Nil(err)
		time.Sleep(time.Millisecond)

		// Make a request that returns no results
		postList, err := s.Store().Post().GetPostsSince(model.GetPostsSinceOptions{ChannelId: channelId, Time: post1.CreateAt}, true)
		s.Require().Nil(err)
		s.Require().Equal(model.NewPostList(), postList)

		// And then ensure that it doesn't cause future requests to also return no results
		postList, err = s.Store().Post().GetPostsSince(model.GetPostsSinceOptions{ChannelId: channelId, Time: post1.CreateAt - 1}, true)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{post1.Id}, postList.Order)

		s.Assert().Len(postList.Posts, 1)
		s.Assert().NotNil(postList.Posts[post1.Id])
	})
}

func (s *PostStoreTestSuite) TestGetPosts() {
	channelId := model.NewId()
	userId := model.NewId()

	post1, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
	})
	s.Require().Nil(err)
	time.Sleep(time.Millisecond)

	post2, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
	})
	s.Require().Nil(err)
	time.Sleep(time.Millisecond)

	post3, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
	})
	s.Require().Nil(err)
	time.Sleep(time.Millisecond)

	post4, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
	})
	s.Require().Nil(err)
	time.Sleep(time.Millisecond)

	post5, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
		RootId:    post3.Id,
	})
	s.Require().Nil(err)
	time.Sleep(time.Millisecond)

	post6, err := s.Store().Post().Save(&model.Post{
		ChannelId: channelId,
		UserId:    userId,
		Message:   "message",
		RootId:    post1.Id,
	})
	s.Require().Nil(err)

	s.T().Run("should return the last posts created in a channel", func(t *testing.T) {
		postList, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: channelId, Page: 0, PerPage: 30, SkipFetchThreads: false}, false)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{
			post6.Id,
			post5.Id,
			post4.Id,
			post3.Id,
			post2.Id,
			post1.Id,
		}, postList.Order)

		s.Assert().Len(postList.Posts, 6)
		s.Assert().NotNil(postList.Posts[post1.Id])
		s.Assert().NotNil(postList.Posts[post2.Id])
		s.Assert().NotNil(postList.Posts[post3.Id])
		s.Assert().NotNil(postList.Posts[post4.Id])
		s.Assert().NotNil(postList.Posts[post5.Id])
		s.Assert().NotNil(postList.Posts[post6.Id])
	})

	s.T().Run("should return the last posts created in a channel and the threads and the reply count must be 0", func(t *testing.T) {
		postList, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: channelId, Page: 0, PerPage: 2, SkipFetchThreads: false}, false)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{
			post6.Id,
			post5.Id,
		}, postList.Order)

		s.Assert().Len(postList.Posts, 4)
		s.Require().NotNil(postList.Posts[post1.Id])
		s.Require().NotNil(postList.Posts[post3.Id])
		s.Require().NotNil(postList.Posts[post5.Id])
		s.Require().NotNil(postList.Posts[post6.Id])
		s.Assert().Equal(int64(0), postList.Posts[post1.Id].ReplyCount)
		s.Assert().Equal(int64(0), postList.Posts[post3.Id].ReplyCount)
		s.Assert().Equal(int64(0), postList.Posts[post5.Id].ReplyCount)
		s.Assert().Equal(int64(0), postList.Posts[post6.Id].ReplyCount)
	})

	s.T().Run("should return the last posts created in a channel without the threads and the reply count must be correct", func(t *testing.T) {
		postList, err := s.Store().Post().GetPosts(model.GetPostsOptions{ChannelId: channelId, Page: 0, PerPage: 2, SkipFetchThreads: true}, false)
		s.Assert().Nil(err)

		s.Assert().Equal([]string{
			post6.Id,
			post5.Id,
		}, postList.Order)

		s.Assert().Len(postList.Posts, 4)
		s.Assert().NotNil(postList.Posts[post5.Id])
		s.Assert().NotNil(postList.Posts[post6.Id])
		s.Assert().Equal(int64(1), postList.Posts[post5.Id].ReplyCount)
		s.Assert().Equal(int64(1), postList.Posts[post6.Id].ReplyCount)
	})
}

func (s *PostStoreTestSuite) TestGetPostBeforeAfter() {
	channelId := model.NewId()

	o0 := &model.Post{}
	o0.ChannelId = channelId
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	_, err := s.Store().Post().Save(o0)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o1 := &model.Post{}
	o1.ChannelId = channelId
	o1.Type = model.POST_JOIN_CHANNEL
	o1.UserId = model.NewId()
	o1.Message = "system_join_channel message"
	_, err = s.Store().Post().Save(o1)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o0a := &model.Post{}
	o0a.ChannelId = channelId
	o0a.UserId = model.NewId()
	o0a.Message = "zz" + model.NewId() + "b"
	o0a.ParentId = o1.Id
	o0a.RootId = o1.Id
	_, err = s.Store().Post().Save(o0a)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o0b := &model.Post{}
	o0b.ChannelId = channelId
	o0b.UserId = model.NewId()
	o0b.Message = "deleted message"
	o0b.ParentId = o1.Id
	o0b.RootId = o1.Id
	o0b.DeleteAt = 1
	_, err = s.Store().Post().Save(o0b)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	otherChannelPost := &model.Post{}
	otherChannelPost.ChannelId = model.NewId()
	otherChannelPost.UserId = model.NewId()
	otherChannelPost.Message = "zz" + model.NewId() + "b"
	_, err = s.Store().Post().Save(otherChannelPost)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = channelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	_, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2a := &model.Post{}
	o2a.ChannelId = channelId
	o2a.UserId = model.NewId()
	o2a.Message = "zz" + model.NewId() + "b"
	o2a.ParentId = o2.Id
	o2a.RootId = o2.Id
	_, err = s.Store().Post().Save(o2a)
	s.Require().Nil(err)

	rPostId1, err := s.Store().Post().GetPostIdBeforeTime(channelId, o0a.CreateAt)
	s.Require().Equal(rPostId1, o1.Id, "should return before post o1")
	s.Require().Nil(err)

	rPostId1, err = s.Store().Post().GetPostIdAfterTime(channelId, o0b.CreateAt)
	s.Require().Equal(rPostId1, o2.Id, "should return before post o2")
	s.Require().Nil(err)

	rPost1, err := s.Store().Post().GetPostAfterTime(channelId, o0b.CreateAt)
	s.Require().Equal(rPost1.Id, o2.Id, "should return before post o2")
	s.Require().Nil(err)

	rPostId2, err := s.Store().Post().GetPostIdBeforeTime(channelId, o0.CreateAt)
	s.Require().Empty(rPostId2, "should return no post")
	s.Require().Nil(err)

	rPostId2, err = s.Store().Post().GetPostIdAfterTime(channelId, o0.CreateAt)
	s.Require().Equal(rPostId2, o1.Id, "should return before post o1")
	s.Require().Nil(err)

	rPost2, err := s.Store().Post().GetPostAfterTime(channelId, o0.CreateAt)
	s.Require().Equal(rPost2.Id, o1.Id, "should return before post o1")
	s.Require().Nil(err)

	rPostId3, err := s.Store().Post().GetPostIdBeforeTime(channelId, o2a.CreateAt)
	s.Require().Equal(rPostId3, o2.Id, "should return before post o2")
	s.Require().Nil(err)

	rPostId3, err = s.Store().Post().GetPostIdAfterTime(channelId, o2a.CreateAt)
	s.Require().Empty(rPostId3, "should return no post")
	s.Require().Nil(err)

	rPost3, err := s.Store().Post().GetPostAfterTime(channelId, o2a.CreateAt)
	s.Require().Empty(rPost3, "should return no post")
	s.Require().Nil(err)
}

func (s *PostStoreTestSuite) TestUserCountsWithPostsByDay() {
	t1 := &model.Team{}
	t1.DisplayName = "DisplayName"
	t1.Name = "zz" + model.NewId() + "b"
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	t1, err := s.Store().Team().Save(t1)
	s.Require().Nil(err)

	c1 := &model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel2"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.CreateAt = utils.MillisFromTime(utils.Yesterday())
	o1.Message = "zz" + model.NewId() + "b"
	o1, nErr = s.Store().Post().Save(o1)
	s.Require().Nil(nErr)

	o1a := &model.Post{}
	o1a.ChannelId = c1.Id
	o1a.UserId = model.NewId()
	o1a.CreateAt = o1.CreateAt
	o1a.Message = "zz" + model.NewId() + "b"
	_, nErr = s.Store().Post().Save(o1a)
	s.Require().Nil(nErr)

	o2 := &model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24)
	o2.Message = "zz" + model.NewId() + "b"
	o2, nErr = s.Store().Post().Save(o2)
	s.Require().Nil(nErr)

	o2a := &model.Post{}
	o2a.ChannelId = c1.Id
	o2a.UserId = o2.UserId
	o2a.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24)
	o2a.Message = "zz" + model.NewId() + "b"
	_, nErr = s.Store().Post().Save(o2a)
	s.Require().Nil(nErr)

	r1, err := s.Store().Post().AnalyticsUserCountsWithPostsByDay(t1.Id)
	s.Require().Nil(err)

	row1 := r1[0]
	s.Require().Equal(float64(2), row1.Value, "wrong value")

	row2 := r1[1]
	s.Require().Equal(float64(1), row2.Value, "wrong value")
}

func (s *PostStoreTestSuite) TestPostCountsByDay() {
	t1 := &model.Team{}
	t1.DisplayName = "DisplayName"
	t1.Name = "zz" + model.NewId() + "b"
	t1.Email = MakeEmail()
	t1.Type = model.TEAM_OPEN
	t1, err := s.Store().Team().Save(t1)
	s.Require().Nil(err)

	c1 := &model.Channel{}
	c1.TeamId = t1.Id
	c1.DisplayName = "Channel2"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, nErr := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(nErr)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.CreateAt = utils.MillisFromTime(utils.Yesterday())
	o1.Message = "zz" + model.NewId() + "b"
	o1, nErr = s.Store().Post().Save(o1)
	s.Require().Nil(nErr)

	o1a := &model.Post{}
	o1a.ChannelId = c1.Id
	o1a.UserId = model.NewId()
	o1a.CreateAt = o1.CreateAt
	o1a.Message = "zz" + model.NewId() + "b"
	_, nErr = s.Store().Post().Save(o1a)
	s.Require().Nil(nErr)

	o2 := &model.Post{}
	o2.ChannelId = c1.Id
	o2.UserId = model.NewId()
	o2.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24 * 2)
	o2.Message = "zz" + model.NewId() + "b"
	o2, nErr = s.Store().Post().Save(o2)
	s.Require().Nil(nErr)

	o2a := &model.Post{}
	o2a.ChannelId = c1.Id
	o2a.UserId = o2.UserId
	o2a.CreateAt = o1.CreateAt - (1000 * 60 * 60 * 24 * 2)
	o2a.Message = "zz" + model.NewId() + "b"
	_, nErr = s.Store().Post().Save(o2a)
	s.Require().Nil(nErr)

	bot1 := &model.Bot{
		Username:    "username",
		Description: "a bot",
		OwnerId:     model.NewId(),
		UserId:      model.NewId(),
	}
	_, nErr = s.Store().Bot().Save(bot1)
	s.Require().Nil(nErr)

	b1 := &model.Post{}
	b1.Message = "bot message one"
	b1.ChannelId = c1.Id
	b1.UserId = bot1.UserId
	b1.CreateAt = utils.MillisFromTime(utils.Yesterday())
	_, nErr = s.Store().Post().Save(b1)
	s.Require().Nil(nErr)

	b1a := &model.Post{}
	b1a.Message = "bot message two"
	b1a.ChannelId = c1.Id
	b1a.UserId = bot1.UserId
	b1a.CreateAt = utils.MillisFromTime(utils.Yesterday()) - (1000 * 60 * 60 * 24 * 2)
	_, nErr = s.Store().Post().Save(b1a)
	s.Require().Nil(nErr)

	time.Sleep(1 * time.Second)

	// summary of posts
	// yesterday - 2 non-bot user posts, 1 bot user post
	// 3 days ago - 2 non-bot user posts, 1 bot user post

	// last 31 days, all users (including bots)
	postCountsOptions := &model.AnalyticsPostCountsOptions{TeamId: t1.Id, BotsOnly: false, YesterdayOnly: false}
	r1, err := s.Store().Post().AnalyticsPostCountsByDay(postCountsOptions)
	s.Require().Nil(err)
	s.Assert().Equal(float64(3), r1[0].Value)
	s.Assert().Equal(float64(3), r1[1].Value)

	// last 31 days, bots only
	postCountsOptions = &model.AnalyticsPostCountsOptions{TeamId: t1.Id, BotsOnly: true, YesterdayOnly: false}
	r1, err = s.Store().Post().AnalyticsPostCountsByDay(postCountsOptions)
	s.Require().Nil(err)
	s.Assert().Equal(float64(1), r1[0].Value)
	s.Assert().Equal(float64(1), r1[1].Value)

	// yesterday only, all users (including bots)
	postCountsOptions = &model.AnalyticsPostCountsOptions{TeamId: t1.Id, BotsOnly: false, YesterdayOnly: true}
	r1, err = s.Store().Post().AnalyticsPostCountsByDay(postCountsOptions)
	s.Require().Nil(err)
	s.Assert().Equal(float64(3), r1[0].Value)

	// yesterday only, bots only
	postCountsOptions = &model.AnalyticsPostCountsOptions{TeamId: t1.Id, BotsOnly: true, YesterdayOnly: true}
	r1, err = s.Store().Post().AnalyticsPostCountsByDay(postCountsOptions)
	s.Require().Nil(err)
	s.Assert().Equal(float64(1), r1[0].Value)

	// total
	r2, err := s.Store().Post().AnalyticsPostCount(t1.Id, false, false)
	s.Require().Nil(err)
	s.Assert().Equal(int64(6), r2)
}

func (s *PostStoreTestSuite) TestGetFlaggedPostsForTeam() {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, err := s.Store().Channel().Save(c1, -1)
	s.Require().Nil(err)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err = s.Store().Post().Save(o1)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4, err = s.Store().Post().Save(o4)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	c2 := &model.Channel{}
	c2.DisplayName = "DMChannel1"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_DIRECT

	m1 := &model.ChannelMember{}
	m1.ChannelId = c2.Id
	m1.UserId = o1.UserId
	m1.NotifyProps = model.GetDefaultChannelNotifyProps()

	m2 := &model.ChannelMember{}
	m2.ChannelId = c2.Id
	m2.UserId = model.NewId()
	m2.NotifyProps = model.GetDefaultChannelNotifyProps()

	c2, err = s.Store().Channel().SaveDirectChannel(c2, m1, m2)
	s.Require().Nil(err)

	o5 := &model.Post{}
	o5.ChannelId = c2.Id
	o5.UserId = m2.UserId
	o5.Message = "zz" + model.NewId() + "b"
	o5, err = s.Store().Post().Save(o5)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	r1, err := s.Store().Post().GetFlaggedPosts(o1.ChannelId, 0, 2)
	s.Require().Nil(err)

	s.Require().Empty(r1.Order, "should be empty")

	preferences := model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o1.Id,
			Value:    "true",
		},
	}

	err = s.Store().Preference().Save(&preferences)
	s.Require().Nil(err)

	r2, err := s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r2.Order, 1, "should have 1 post")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o2.Id,
			Value:    "true",
		},
	}

	err = s.Store().Preference().Save(&preferences)
	s.Require().Nil(err)

	r3, err := s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 1)
	s.Require().Nil(err)
	s.Require().Len(r3.Order, 1, "should have 1 post")

	r3, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 1, 1)
	s.Require().Nil(err)
	s.Require().Len(r3.Order, 1, "should have 1 post")

	r3, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 1000, 10)
	s.Require().Nil(err)
	s.Require().Empty(r3.Order, "should be empty")

	r4, err := s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 2, "should have 2 posts")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o3.Id,
			Value:    "true",
		},
	}

	err = s.Store().Preference().Save(&preferences)
	s.Require().Nil(err)

	r4, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 2, "should have 2 posts")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o4.Id,
			Value:    "true",
		},
	}
	err = s.Store().Preference().Save(&preferences)
	s.Require().Nil(err)

	r4, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 2, "should have 2 posts")

	r4, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, model.NewId(), 0, 2)
	s.Require().Nil(err)
	s.Require().Empty(r4.Order, "should have 0 posts")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o5.Id,
			Value:    "true",
		},
	}
	err = s.Store().Preference().Save(&preferences)
	s.Require().Nil(err)

	r4, err = s.Store().Post().GetFlaggedPostsForTeam(o1.UserId, c1.TeamId, 0, 10)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 3, "should have 3 posts")

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *PostStoreTestSuite) TestGetFlaggedPosts() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	r1, err := s.Store().Post().GetFlaggedPosts(o1.UserId, 0, 2)
	s.Require().Nil(err)
	s.Require().Empty(r1.Order, "should be empty")

	preferences := model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o1.Id,
			Value:    "true",
		},
	}

	nErr := s.Store().Preference().Save(&preferences)
	s.Require().Nil(nErr)

	r2, err := s.Store().Post().GetFlaggedPosts(o1.UserId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r2.Order, 1, "should have 1 post")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o2.Id,
			Value:    "true",
		},
	}

	nErr = s.Store().Preference().Save(&preferences)
	s.Require().Nil(nErr)

	r3, err := s.Store().Post().GetFlaggedPosts(o1.UserId, 0, 1)
	s.Require().Nil(err)
	s.Require().Len(r3.Order, 1, "should have 1 post")

	r3, err = s.Store().Post().GetFlaggedPosts(o1.UserId, 1, 1)
	s.Require().Nil(err)
	s.Require().Len(r3.Order, 1, "should have 1 post")

	r3, err = s.Store().Post().GetFlaggedPosts(o1.UserId, 1000, 10)
	s.Require().Nil(err)
	s.Require().Empty(r3.Order, "should be empty")

	r4, err := s.Store().Post().GetFlaggedPosts(o1.UserId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 2, "should have 2 posts")

	preferences = model.Preferences{
		{
			UserId:   o1.UserId,
			Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
			Name:     o3.Id,
			Value:    "true",
		},
	}

	nErr = s.Store().Preference().Save(&preferences)
	s.Require().Nil(nErr)

	r4, err = s.Store().Post().GetFlaggedPosts(o1.UserId, 0, 2)
	s.Require().Nil(err)
	s.Require().Len(r4.Order, 2, "should have 2 posts")
}

func (s *PostStoreTestSuite) TestGetFlaggedPostsForChannel() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	// deleted post
	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = o1.ChannelId
	o3.Message = "zz" + model.NewId() + "b"
	o3.DeleteAt = 1
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	o4 := &model.Post{}
	o4.ChannelId = model.NewId()
	o4.UserId = model.NewId()
	o4.Message = "zz" + model.NewId() + "b"
	o4, err = s.Store().Post().Save(o4)
	s.Require().Nil(err)
	time.Sleep(2 * time.Millisecond)

	r, err := s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)
	s.Require().Nil(err)
	s.Require().Empty(r.Order, "should be empty")

	preference := model.Preference{
		UserId:   o1.UserId,
		Category: model.PREFERENCE_CATEGORY_FLAGGED_POST,
		Name:     o1.Id,
		Value:    "true",
	}

	nErr := s.Store().Preference().Save(&model.Preferences{preference})
	s.Require().Nil(nErr)

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)
	s.Require().Nil(err)
	s.Require().Len(r.Order, 1, "should have 1 post")

	preference.Name = o2.Id
	nErr = s.Store().Preference().Save(&model.Preferences{preference})
	s.Require().Nil(nErr)

	preference.Name = o3.Id
	nErr = s.Store().Preference().Save(&model.Preferences{preference})
	s.Require().Nil(nErr)

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 1)
	s.Require().Nil(err)
	s.Require().Len(r.Order, 1, "should have 1 post")

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 1, 1)
	s.Require().Nil(err)
	s.Require().Len(r.Order, 1, "should have 1 post")

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 1000, 10)
	s.Require().Nil(err)
	s.Require().Empty(r.Order, "should be empty")

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o1.ChannelId, 0, 10)
	s.Require().Nil(err)
	s.Require().Len(r.Order, 2, "should have 2 posts")

	preference.Name = o4.Id
	nErr = s.Store().Preference().Save(&model.Preferences{preference})
	s.Require().Nil(nErr)

	r, err = s.Store().Post().GetFlaggedPostsForChannel(o1.UserId, o4.ChannelId, 0, 10)
	s.Require().Nil(err)
	s.Require().Len(r.Order, 1, "should have 1 posts")
}

func (s *PostStoreTestSuite) TestGetPostsCreatedAt() {
	createTime := model.GetMillis() + 1

	o0 := &model.Post{}
	o0.ChannelId = model.NewId()
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	o0.CreateAt = createTime
	o0, err := s.Store().Post().Save(o0)
	s.Require().Nil(err)

	o1 := &model.Post{}
	o1.ChannelId = o0.ChannelId
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1.CreateAt = createTime
	o1, err = s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2.CreateAt = createTime + 1
	_, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "b"
	o3.CreateAt = createTime
	_, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	r1, _ := s.Store().Post().GetPostsCreatedAt(o1.ChannelId, createTime)
	s.Assert().Equal(2, len(r1))
}

func (s *PostStoreTestSuite) TestOverwriteMultiple() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	o4, err := s.Store().Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test"},
	})
	s.Require().Nil(err)

	o5, err := s.Store().Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test2", "test3"},
	})
	s.Require().Nil(err)

	r1, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro1 := r1.Posts[o1.Id]

	r2, err := s.Store().Post().Get(o2.Id, false)
	s.Require().Nil(err)
	ro2 := r2.Posts[o2.Id]

	r3, err := s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err)
	ro3 := r3.Posts[o3.Id]

	r4, err := s.Store().Post().Get(o4.Id, false)
	s.Require().Nil(err)
	ro4 := r4.Posts[o4.Id]

	r5, err := s.Store().Post().Get(o5.Id, false)
	s.Require().Nil(err)
	ro5 := r5.Posts[o5.Id]

	s.Require().Equal(ro1.Message, o1.Message, "Failed to save/get")
	s.Require().Equal(ro2.Message, o2.Message, "Failed to save/get")
	s.Require().Equal(ro3.Message, o3.Message, "Failed to save/get")
	s.Require().Equal(ro4.Message, o4.Message, "Failed to save/get")
	s.Require().Equal(ro4.Filenames, o4.Filenames, "Failed to save/get")
	s.Require().Equal(ro5.Message, o5.Message, "Failed to save/get")
	s.Require().Equal(ro5.Filenames, o5.Filenames, "Failed to save/get")

	s.T().Run("overwrite changing message", func(t *testing.T) {
		o1a := ro1.Clone()
		o1a.Message = ro1.Message + "BBBBBBBBBB"

		o2a := ro2.Clone()
		o2a.Message = ro2.Message + "DDDDDDD"

		o3a := ro3.Clone()
		o3a.Message = ro3.Message + "WWWWWWW"

		_, errIdx, err := s.Store().Post().OverwriteMultiple([]*model.Post{o1a, o2a, o3a})
		s.Require().Nil(err)
		s.Require().Equal(-1, errIdx)

		r1, nErr := s.Store().Post().Get(o1.Id, false)
		s.Require().Nil(nErr)
		ro1a := r1.Posts[o1.Id]

		r2, nErr = s.Store().Post().Get(o1.Id, false)
		s.Require().Nil(nErr)
		ro2a := r2.Posts[o2.Id]

		r3, nErr = s.Store().Post().Get(o3.Id, false)
		s.Require().Nil(nErr)
		ro3a := r3.Posts[o3.Id]

		s.Assert().Equal(ro1a.Message, o1a.Message, "Failed to overwrite/get")
		s.Assert().Equal(ro2a.Message, o2a.Message, "Failed to overwrite/get")
		s.Assert().Equal(ro3a.Message, o3a.Message, "Failed to overwrite/get")
	})

	s.T().Run("overwrite clearing filenames", func(t *testing.T) {
		o4a := ro4.Clone()
		o4a.Filenames = []string{}
		o4a.FileIds = []string{model.NewId()}

		o5a := ro5.Clone()
		o5a.Filenames = []string{}
		o5a.FileIds = []string{}

		_, errIdx, err := s.Store().Post().OverwriteMultiple([]*model.Post{o4a, o5a})
		s.Require().Nil(err)
		s.Require().Equal(-1, errIdx)

		r4, nErr := s.Store().Post().Get(o4.Id, false)
		s.Require().Nil(nErr)
		ro4a := r4.Posts[o4.Id]

		r5, nErr = s.Store().Post().Get(o5.Id, false)
		s.Require().Nil(nErr)
		ro5a := r5.Posts[o5.Id]

		s.Require().Empty(ro4a.Filenames, "Failed to clear Filenames")
		s.Require().Len(ro4a.FileIds, 1, "Failed to set FileIds")
		s.Require().Empty(ro5a.Filenames, "Failed to clear Filenames")
		s.Require().Empty(ro5a.FileIds, "Failed to set FileIds")
	})
}

func (s *PostStoreTestSuite) TestOverwrite() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2.ParentId = o1.Id
	o2.RootId = o1.Id
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	o4, err := s.Store().Post().Save(&model.Post{
		ChannelId: model.NewId(),
		UserId:    model.NewId(),
		Message:   model.NewId(),
		Filenames: []string{"test"},
	})
	s.Require().Nil(err)

	r1, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro1 := r1.Posts[o1.Id]

	r2, err := s.Store().Post().Get(o2.Id, false)
	s.Require().Nil(err)
	ro2 := r2.Posts[o2.Id]

	r3, err := s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err)
	ro3 := r3.Posts[o3.Id]

	r4, err := s.Store().Post().Get(o4.Id, false)
	s.Require().Nil(err)
	ro4 := r4.Posts[o4.Id]

	s.Require().Equal(ro1.Message, o1.Message, "Failed to save/get")
	s.Require().Equal(ro2.Message, o2.Message, "Failed to save/get")
	s.Require().Equal(ro3.Message, o3.Message, "Failed to save/get")
	s.Require().Equal(ro4.Message, o4.Message, "Failed to save/get")

	s.T().Run("overwrite changing message", func(t *testing.T) {
		o1a := ro1.Clone()
		o1a.Message = ro1.Message + "BBBBBBBBBB"
		_, err = s.Store().Post().Overwrite(o1a)
		s.Require().Nil(err)

		o2a := ro2.Clone()
		o2a.Message = ro2.Message + "DDDDDDD"
		_, err = s.Store().Post().Overwrite(o2a)
		s.Require().Nil(err)

		o3a := ro3.Clone()
		o3a.Message = ro3.Message + "WWWWWWW"
		_, err = s.Store().Post().Overwrite(o3a)
		s.Require().Nil(err)

		r1, err = s.Store().Post().Get(o1.Id, false)
		s.Require().Nil(err)
		ro1a := r1.Posts[o1.Id]

		r2, err = s.Store().Post().Get(o1.Id, false)
		s.Require().Nil(err)
		ro2a := r2.Posts[o2.Id]

		r3, err = s.Store().Post().Get(o3.Id, false)
		s.Require().Nil(err)
		ro3a := r3.Posts[o3.Id]

		s.Assert().Equal(ro1a.Message, o1a.Message, "Failed to overwrite/get")
		s.Assert().Equal(ro2a.Message, o2a.Message, "Failed to overwrite/get")
		s.Assert().Equal(ro3a.Message, o3a.Message, "Failed to overwrite/get")
	})

	s.T().Run("overwrite clearing filenames", func(t *testing.T) {
		o4a := ro4.Clone()
		o4a.Filenames = []string{}
		o4a.FileIds = []string{model.NewId()}
		_, err = s.Store().Post().Overwrite(o4a)
		s.Require().Nil(err)

		r4, err = s.Store().Post().Get(o4.Id, false)
		s.Require().Nil(err)

		ro4a := r4.Posts[o4.Id]
		s.Require().Empty(ro4a.Filenames, "Failed to clear Filenames")
		s.Require().Len(ro4a.FileIds, 1, "Failed to set FileIds")
	})
}

func (s *PostStoreTestSuite) TestGetPostsByIds() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = o1.ChannelId
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	r1, err := s.Store().Post().Get(o1.Id, false)
	s.Require().Nil(err)
	ro1 := r1.Posts[o1.Id]

	r2, err := s.Store().Post().Get(o2.Id, false)
	s.Require().Nil(err)
	ro2 := r2.Posts[o2.Id]

	r3, err := s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err)
	ro3 := r3.Posts[o3.Id]

	postIds := []string{
		ro1.Id,
		ro2.Id,
		ro3.Id,
	}

	posts, err := s.Store().Post().GetPostsByIds(postIds)
	s.Require().Nil(err)
	s.Require().Len(posts, 3, "Expected 3 posts in results. Got %v", len(posts))

	err = s.Store().Post().Delete(ro1.Id, model.GetMillis(), "")
	s.Require().Nil(err)

	posts, err = s.Store().Post().GetPostsByIds(postIds)
	s.Require().Nil(err)
	s.Require().Len(posts, 3, "Expected 3 posts in results. Got %v", len(posts))
}

func (s *PostStoreTestSuite) TestGetPostsBatchForIndexing() {
	c1 := &model.Channel{}
	c1.TeamId = model.NewId()
	c1.DisplayName = "Channel1"
	c1.Name = "zz" + model.NewId() + "b"
	c1.Type = model.CHANNEL_OPEN
	c1, _ = s.Store().Channel().Save(c1, -1)

	c2 := &model.Channel{}
	c2.TeamId = model.NewId()
	c2.DisplayName = "Channel2"
	c2.Name = "zz" + model.NewId() + "b"
	c2.Type = model.CHANNEL_OPEN
	c2, _ = s.Store().Channel().Save(c2, -1)

	o1 := &model.Post{}
	o1.ChannelId = c1.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = c2.Id
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "CCCCCCCCC"
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = c1.Id
	o3.UserId = model.NewId()
	o3.ParentId = o1.Id
	o3.RootId = o1.Id
	o3.Message = "zz" + model.NewId() + "QQQQQQQQQQ"
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	r, err := s.Store().Post().GetPostsBatchForIndexing(o1.CreateAt, model.GetMillis()+100000, 100)
	s.Require().Nil(err)
	s.Require().Len(r, 3, "Expected 3 posts in results. Got %v", len(r))
	for _, p := range r {
		if p.Id == o1.Id {
			s.Require().Equal(p.TeamId, c1.TeamId, "Unexpected team ID")
			s.Require().Nil(p.ParentCreateAt, "Unexpected parent create at")
		} else if p.Id == o2.Id {
			s.Require().Equal(p.TeamId, c2.TeamId, "Unexpected team ID")
			s.Require().Nil(p.ParentCreateAt, "Unexpected parent create at")
		} else if p.Id == o3.Id {
			s.Require().Equal(p.TeamId, c1.TeamId, "Unexpected team ID")
			s.Require().Equal(*p.ParentCreateAt, o1.CreateAt, "Unexpected parent create at")
		} else {
			s.Require().Fail("unexpected post returned")
		}
	}
}

func (s *PostStoreTestSuite) TestPermanentDeleteBatch() {
	o1 := &model.Post{}
	o1.ChannelId = model.NewId()
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o1.CreateAt = 1000
	o1, err := s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = model.NewId()
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o2.CreateAt = 1000
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	o3 := &model.Post{}
	o3.ChannelId = model.NewId()
	o3.UserId = model.NewId()
	o3.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	o3.CreateAt = 100000
	o3, err = s.Store().Post().Save(o3)
	s.Require().Nil(err)

	_, err = s.Store().Post().PermanentDeleteBatch(2000, 1000)
	s.Require().Nil(err)

	_, err = s.Store().Post().Get(o1.Id, false)
	s.Require().NotNil(err, "Should have not found post 1 after purge")

	_, err = s.Store().Post().Get(o2.Id, false)
	s.Require().NotNil(err, "Should have not found post 2 after purge")

	_, err = s.Store().Post().Get(o3.Id, false)
	s.Require().Nil(err, "Should have not found post 3 after purge")
}

func (s *PostStoreTestSuite) TestGetOldest() {
	o0 := &model.Post{}
	o0.ChannelId = model.NewId()
	o0.UserId = model.NewId()
	o0.Message = "zz" + model.NewId() + "b"
	o0.CreateAt = 3
	o0, err := s.Store().Post().Save(o0)
	s.Require().Nil(err)

	o1 := &model.Post{}
	o1.ChannelId = o0.Id
	o1.UserId = model.NewId()
	o1.Message = "zz" + model.NewId() + "b"
	o1.CreateAt = 2
	o1, err = s.Store().Post().Save(o1)
	s.Require().Nil(err)

	o2 := &model.Post{}
	o2.ChannelId = o1.ChannelId
	o2.UserId = model.NewId()
	o2.Message = "zz" + model.NewId() + "b"
	o2.CreateAt = 1
	o2, err = s.Store().Post().Save(o2)
	s.Require().Nil(err)

	r1, err := s.Store().Post().GetOldest()

	s.Require().Nil(err)
	s.Assert().EqualValues(o2.Id, r1.Id)
}

func (s *PostStoreTestSuite) TestGetMaxPostSize() {
	s.Assert().Equal(model.POST_MESSAGE_MAX_RUNES_V2, s.Store().Post().GetMaxPostSize())
	s.Assert().Equal(model.POST_MESSAGE_MAX_RUNES_V2, s.Store().Post().GetMaxPostSize())
}

func (s *PostStoreTestSuite) TestGetParentsForExportAfter() {
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
	u1.Username = model.NewId()
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err = s.Store().User().Save(&u1)
	s.Require().Nil(err)

	p1 := &model.Post{}
	p1.ChannelId = c1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1, nErr = s.Store().Post().Save(p1)
	s.Require().Nil(nErr)

	posts, err := s.Store().Post().GetParentsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(err)

	found := false
	for _, p := range posts {
		if p.Id == p1.Id {
			found = true
			s.Assert().Equal(p.Id, p1.Id)
			s.Assert().Equal(p.Message, p1.Message)
			s.Assert().Equal(p.Username, u1.Username)
			s.Assert().Equal(p.TeamName, t1.Name)
			s.Assert().Equal(p.ChannelName, c1.Name)
		}
	}
	s.Assert().True(found)
}

func (s *PostStoreTestSuite) TestGetRepliesForExport() {
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

	p1 := &model.Post{}
	p1.ChannelId = c1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1, nErr = s.Store().Post().Save(p1)
	s.Require().Nil(nErr)

	p2 := &model.Post{}
	p2.ChannelId = c1.Id
	p2.UserId = u1.Id
	p2.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p2.CreateAt = 1001
	p2.ParentId = p1.Id
	p2.RootId = p1.Id
	p2, nErr = s.Store().Post().Save(p2)
	s.Require().Nil(nErr)

	r1, err := s.Store().Post().GetRepliesForExport(p1.Id)
	s.Assert().Nil(err)

	s.Assert().Len(r1, 1)

	reply1 := r1[0]
	s.Assert().Equal(reply1.Id, p2.Id)
	s.Assert().Equal(reply1.Message, p2.Message)
	s.Assert().Equal(reply1.Username, u1.Username)

	// Checking whether replies by deleted user are exported
	u1.DeleteAt = 1002
	_, err = s.Store().User().Update(&u1, false)
	s.Require().Nil(err)

	r1, err = s.Store().Post().GetRepliesForExport(p1.Id)
	s.Assert().Nil(err)

	s.Assert().Len(r1, 1)

	reply1 = r1[0]
	s.Assert().Equal(reply1.Id, p2.Id)
	s.Assert().Equal(reply1.Message, p2.Message)
	s.Assert().Equal(reply1.Username, u1.Username)

}

func (s *PostStoreTestSuite) TestGetDirectPostParentsForExportAfter() {
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

	s.Store().Channel().SaveDirectChannel(&o1, &m1, &m2)

	p1 := &model.Post{}
	p1.ChannelId = o1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
	p1.CreateAt = 1000
	p1, nErr = s.Store().Post().Save(p1)
	s.Require().Nil(nErr)

	r1, nErr := s.Store().Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(nErr)

	s.Assert().Equal(p1.Message, r1[0].Message)

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *PostStoreTestSuite) TestGetDirectPostParentsForExportAfterDeleted() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	u1 := &model.User{}
	u1.DeleteAt = 1
	u1.Email = MakeEmail()
	u1.Nickname = model.NewId()
	_, err := s.Store().User().Save(u1)
	s.Require().Nil(err)
	_, nErr := s.Store().Team().SaveMember(&model.TeamMember{TeamId: model.NewId(), UserId: u1.Id}, -1)
	s.Require().Nil(nErr)

	u2 := &model.User{}
	u2.DeleteAt = 1
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
	s.Assert().Nil(nErr)

	p1 := &model.Post{}
	p1.ChannelId = o1.Id
	p1.UserId = u1.Id
	p1.Message = "zz" + model.NewId() + "BBBBBBBBBBBB"
	p1.CreateAt = 1000
	p1, nErr = s.Store().Post().Save(p1)
	s.Require().Nil(nErr)

	o1a := p1.Clone()
	o1a.DeleteAt = 1
	o1a.Message = p1.Message + "BBBBBBBBBB"
	_, nErr = s.Store().Post().Update(o1a, p1)
	s.Require().Nil(nErr)

	r1, nErr := s.Store().Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(nErr)

	s.Assert().Equal(0, len(r1))

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}

func (s *PostStoreTestSuite) TestGetDirectPostParentsForExportAfterBatched() {
	teamId := model.NewId()

	o1 := model.Channel{}
	o1.TeamId = teamId
	o1.DisplayName = "Name"
	o1.Name = "zz" + model.NewId() + "b"
	o1.Type = model.CHANNEL_DIRECT

	var postIds []string
	for i := 0; i < 150; i++ {
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

		p1 := &model.Post{}
		p1.ChannelId = o1.Id
		p1.UserId = u1.Id
		p1.Message = "zz" + model.NewId() + "AAAAAAAAAAA"
		p1.CreateAt = 1000
		p1, nErr = s.Store().Post().Save(p1)
		s.Require().Nil(nErr)
		postIds = append(postIds, p1.Id)
	}
	sort.Slice(postIds, func(i, j int) bool { return postIds[i] < postIds[j] })

	// Get all posts
	r1, err := s.Store().Post().GetDirectPostParentsForExportAfter(10000, strings.Repeat("0", 26))
	s.Assert().Nil(err)
	s.Assert().Equal(len(postIds), len(r1))
	var exportedPostIds []string
	for i := range r1 {
		exportedPostIds = append(exportedPostIds, r1[i].Id)
	}
	sort.Slice(exportedPostIds, func(i, j int) bool { return exportedPostIds[i] < exportedPostIds[j] })
	s.Assert().ElementsMatch(postIds, exportedPostIds)

	// Get 100
	r1, err = s.Store().Post().GetDirectPostParentsForExportAfter(100, strings.Repeat("0", 26))
	s.Assert().Nil(err)
	s.Assert().Equal(100, len(r1))
	exportedPostIds = []string{}
	for i := range r1 {
		exportedPostIds = append(exportedPostIds, r1[i].Id)
	}
	sort.Slice(exportedPostIds, func(i, j int) bool { return exportedPostIds[i] < exportedPostIds[j] })
	s.Assert().ElementsMatch(postIds[:100], exportedPostIds)

	// Manually truncate Channels table until testlib can handle cleanups
	s.SqlStore().GetMaster().Exec("TRUNCATE Channels")
}
