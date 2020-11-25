package sqlstore

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/mlog"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
	"sync"
)

type SqlPostStore struct {
	*SqlSupplier
	metrics           einterfaces.MetricsInterface
	maxPostSizeOnce   sync.Once
	maxPostSizeCached int
}

func (s *SqlPostStore) ClearCaches() {
}

func postSliceColumns() []string {
	return []string{"Id", "CreateAt", "UpdateAt", "EditAt", "DeleteAt", "IsPinned", "UserId", "ChannelId", "RootId", "ParentId", "OriginalId", "Message", "Type", "Props", "Hashtags", "Filenames", "FileIds", "HasReactions"}
}

func postToSlice(post *model.Post) []interface{} {
	return []interface{}{
		post.Id,
		post.CreateAt,
		post.UpdateAt,
		post.EditAt,
		post.DeleteAt,
		post.IsPinned,
		post.UserId,
		post.ChannelId,
		post.RootId,
		post.ParentId,
		post.OriginalId,
		post.Message,
		post.Type,
		model.StringInterfaceToJson(post.Props),
		post.Hashtags,
		model.ArrayToJson(post.Filenames),
		model.ArrayToJson(post.FileIds),
		post.HasReactions,
	}
}

func newSqlPostStore(sqlSupplier *SqlSupplier, metrics einterfaces.MetricsInterface) store.PostStore {
	s := &SqlPostStore{
		SqlSupplier:       sqlSupplier,
		metrics:           metrics,
		maxPostSizeCached: model.POST_MESSAGE_MAX_RUNES_V1,
	}

	for _, db := range sqlSupplier.GetAllConns() {
		table := db.AddTableWithName(model.Post{}, "Posts").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("RootId").SetMaxSize(26)
		table.ColMap("ParentId").SetMaxSize(26)
		table.ColMap("OriginalId").SetMaxSize(26)
		table.ColMap("Message").SetMaxSize(model.POST_MESSAGE_MAX_BYTES_V2)
		table.ColMap("Type").SetMaxSize(26)
		table.ColMap("Hashtags").SetMaxSize(1000)
		table.ColMap("Props").SetMaxSize(8000)
		table.ColMap("Filenames").SetMaxSize(model.POST_FILENAMES_MAX_RUNES)
		table.ColMap("FileIds").SetMaxSize(150)
	}

	return s
}

func (s *SqlPostStore) SaveMultiple(posts []*model.Post) ([]*model.Post, int, error) {
	channelNewPosts := make(map[string]int)
	maxDateNewPosts := make(map[string]int64)
	rootIds := make(map[string]int)
	maxDateRootIds := make(map[string]int64)
	for idx, post := range posts {
		if len(post.Id) > 0 {
			return nil, idx, store.NewErrInvalidInput("Post", "id", post.Id)
		}
		post.PreSave()
		maxPostSize := s.GetMaxPostSize()
		if err := post.IsValid(maxPostSize); err != nil {
			return nil, idx, err
		}

		currentChannelCount, ok := channelNewPosts[post.ChannelId]
		if !ok {
			if post.IsJoinLeaveMessage() {
				channelNewPosts[post.ChannelId] = 0
			} else {
				channelNewPosts[post.ChannelId] = 1
			}
			maxDateNewPosts[post.ChannelId] = post.CreateAt
		} else {
			if !post.IsJoinLeaveMessage() {
				channelNewPosts[post.ChannelId] = currentChannelCount + 1
			}
			if post.CreateAt > maxDateNewPosts[post.ChannelId] {
				maxDateNewPosts[post.ChannelId] = post.CreateAt
			}
		}

		if len(post.RootId) == 0 {
			continue
		}

		currentRootCount, ok := rootIds[post.RootId]
		if !ok {
			rootIds[post.RootId] = 1
			maxDateRootIds[post.RootId] = post.CreateAt
		} else {
			rootIds[post.RootId] = currentRootCount + 1
			if post.CreateAt > maxDateRootIds[post.RootId] {
				maxDateRootIds[post.RootId] = post.CreateAt
			}
		}
	}

	builder := s.getQueryBuilder().Insert("Posts").Columns(postSliceColumns()...)
	for _, post := range posts {
		builder = builder.Values(postToSlice(post)...)
	}
	query, args, err := builder.ToSql()
	if err != nil {
		return nil, -1, errors.Wrap(err, "post_tosql")
	}

	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return posts, -1, errors.Wrap(err, "begin_transaction")
	}

	defer finalizeTransaction(transaction)

	if _, err = transaction.Exec(query, args...); err != nil {
		return nil, -1, errors.Wrap(err, "failed to save Post")
	}

	if err = s.updateThreadsFromPosts(transaction, posts); err != nil {
		mlog.Error("Error updating posts, thread update failed", mlog.Err(err))
	}

	if err = transaction.Commit(); err != nil {
		// don't need to rollback here since the transaction is already closed
		return posts, -1, errors.Wrap(err, "commit_transaction")
	}

	for channelId, count := range channelNewPosts {
		if _, err = s.GetMaster().Exec("UPDATE Channels SET LastPostAt = GREATEST(:LastPostAt, LastPostAt), TotalMsgCount = TotalMsgCount + :Count WHERE Id = :ChannelId", map[string]interface{}{"LastPostAt": maxDateNewPosts[channelId], "ChannelId": channelId, "Count": count}); err != nil {
			mlog.Error("Error updating Channel LastPostAt.", mlog.Err(err))
		}
	}

	for rootId := range rootIds {
		if _, err = s.GetMaster().Exec("UPDATE Posts SET UpdateAt = :UpdateAt WHERE Id = :RootId", map[string]interface{}{"UpdateAt": maxDateRootIds[rootId], "RootId": rootId}); err != nil {
			mlog.Error("Error updating Post UpdateAt.", mlog.Err(err))
		}
	}

	unknownRepliesPosts := []*model.Post{}
	for _, post := range posts {
		if len(post.RootId) == 0 {
			count, ok := rootIds[post.Id]
			if ok {
				post.ReplyCount += int64(count)
			}
		} else {
			unknownRepliesPosts = append(unknownRepliesPosts, post)
		}
	}

	if len(unknownRepliesPosts) > 0 {
		if err := s.populateReplyCount(unknownRepliesPosts); err != nil {
			mlog.Error("Unable to populate the reply count in some posts.", mlog.Err(err))
		}
	}

	return posts, -1, nil
}

func (s *SqlPostStore) populateReplyCount(posts []*model.Post) error {
	rootIds := []string{}
	for _, post := range posts {
		rootIds = append(rootIds, post.RootId)
	}
	countList := []struct {
		RootId string
		Count  int64
	}{}
	query := s.getQueryBuilder().Select("RootId, COUNT(Id) AS Count").From("Posts").Where(sq.Eq{"RootId": rootIds}).Where(sq.Eq{"DeleteAt": 0}).GroupBy("RootId")

	queryString, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "post_tosql")
	}
	_, err = s.GetMaster().Select(&countList, queryString, args...)
	if err != nil {
		return errors.Wrap(err, "failed to count Posts")
	}

	counts := map[string]int64{}
	for _, count := range countList {
		counts[count.RootId] = count.Count
	}

	for _, post := range posts {
		count, ok := counts[post.RootId]
		if !ok {
			post.ReplyCount = 0
		}
		post.ReplyCount = count
	}

	return nil
}

func (s *SqlPostStore) Save(post *model.Post) (*model.Post, error) {
	posts, _, err := s.SaveMultiple([]*model.Post{post})
	if err != nil {
		return nil, err
	}
	return posts[0], nil
}

func (s *SqlPostStore) Get(id string, skipFetchThreads bool) (*model.PostList, error) {
	pl := model.NewPostList()

	if len(id) == 0 {
		return nil, store.NewErrInvalidInput("Post", "id", id)
	}

	var post model.Post
	postFetchQuery := "SELECT p.*, (SELECT count(Posts.Id) FROM Posts WHERE Posts.RootId = (CASE WHEN p.RootId = '' THEN p.Id ELSE p.RootId END) AND Posts.DeleteAt = 0) as ReplyCount FROM Posts p WHERE p.Id = :Id AND p.DeleteAt = 0"
	err := s.GetReplica().SelectOne(&post, postFetchQuery, map[string]interface{}{"Id": id})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Post", id)
		}

		return nil, errors.Wrapf(err, "failed to get Post with id=%s", id)
	}
	pl.AddPost(&post)
	pl.AddOrder(id)
	if !skipFetchThreads {
		rootId := post.RootId

		if rootId == "" {
			rootId = post.Id
		}

		if len(rootId) == 0 {
			return nil, errors.Wrapf(err, "invalid rootId with value=%s", rootId)
		}

		var posts []*model.Post
		_, err = s.GetReplica().Select(&posts, "SELECT *, (SELECT count(Id) FROM Posts WHERE Posts.RootId = (CASE WHEN p.RootId = '' THEN p.Id ELSE p.RootId END) AND Posts.DeleteAt = 0) as ReplyCount FROM Posts p WHERE (Id = :Id OR RootId = :RootId) AND DeleteAt = 0", map[string]interface{}{"Id": rootId, "RootId": rootId})
		if err != nil {
			return nil, errors.Wrap(err, "failed to find Posts")
		}

		for _, p := range posts {
			pl.AddPost(p)
			pl.AddOrder(p.Id)
		}
	}
	return pl, nil
}

func (s *SqlPostStore) GetSingle(id string) (*model.Post, error) {
	var post model.Post
	err := s.GetReplica().SelectOne(&post, "SELECT * FROM Posts WHERE Id = :Id AND DeleteAt = 0", map[string]interface{}{"Id": id})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Post", id)
		}

		return nil, errors.Wrapf(err, "failed to get Post with id=%s", id)
	}
	return &post, nil
}

func (s *SqlPostStore) determineMaxPostSize() int {
	var maxPostSizeBytes int32

	if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		// The Post.Message column in Postgres has historically been VARCHAR(4000), but
		// may be manually enlarged to support longer posts.
		if err := s.GetReplica().SelectOne(&maxPostSizeBytes, `
			SELECT
				COALESCE(character_maximum_length, 0)
			FROM
				information_schema.columns
			WHERE
				table_name = 'posts'
			AND	column_name = 'message'
		`); err != nil {
			mlog.Warn("Unable to determine the maximum supported post size", mlog.Err(err))
		}
	} else if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		// The Post.Message column in MySQL has historically been TEXT, with a maximum
		// limit of 65535.
		if err := s.GetReplica().SelectOne(&maxPostSizeBytes, `
			SELECT
				COALESCE(CHARACTER_MAXIMUM_LENGTH, 0)
			FROM
				INFORMATION_SCHEMA.COLUMNS
			WHERE
				table_schema = DATABASE()
			AND	table_name = 'Posts'
			AND	column_name = 'Message'
			LIMIT 0, 1
		`); err != nil {
			mlog.Error("Unable to determine the maximum supported post size", mlog.Err(err))
		}
	} else {
		mlog.Warn("No implementation found to determine the maximum supported post size")
	}

	// Assume a worst-case representation of four bytes per rune.
	maxPostSize := int(maxPostSizeBytes) / 4

	// To maintain backwards compatibility, don't yield a maximum post
	// size smaller than the previous limit, even though it wasn't
	// actually possible to store 4000 runes in all cases.
	if maxPostSize < model.POST_MESSAGE_MAX_RUNES_V1 {
		maxPostSize = model.POST_MESSAGE_MAX_RUNES_V1
	}

	mlog.Info("Post.Message has size restrictions", mlog.Int("max_characters", maxPostSize), mlog.Int32("max_bytes", maxPostSizeBytes))

	return maxPostSize
}

// GetMaxPostSize returns the maximum number of runes that may be stored in a post.
func (s *SqlPostStore) GetMaxPostSize() int {
	s.maxPostSizeOnce.Do(func() {
		s.maxPostSizeCached = s.determineMaxPostSize()
	})
	return s.maxPostSizeCached
}

func (s *SqlPostStore) updateThreadsFromPosts(transaction *gorp.Transaction, posts []*model.Post) error {
	postsByRoot := map[string][]*model.Post{}
	var rootIds []string
	for _, post := range posts {
		// skip if post is not a part of a thread
		if len(post.RootId) == 0 {
			continue
		}
		rootIds = append(rootIds, post.RootId)
		postsByRoot[post.RootId] = append(postsByRoot[post.RootId], post)
	}
	if len(rootIds) == 0 {
		return nil
	}
	now := model.GetMillis()
	threadsByRootsSql, threadsByRootsArgs, _ := s.getQueryBuilder().Select("*").From("Threads").Where(sq.Eq{"PostId": rootIds}).ToSql()
	var threadsByRoots []*model.Thread
	if _, err := transaction.Select(&threadsByRoots, threadsByRootsSql, threadsByRootsArgs...); err != nil {
		return err
	}

	threadByRoot := map[string]*model.Thread{}
	for _, thread := range threadsByRoots {
		threadByRoot[thread.PostId] = thread
	}

	for rootId, posts := range postsByRoot {
		if thread, found := threadByRoot[rootId]; !found {
			// calculate participants
			var participants model.StringArray
			if _, err := transaction.Select(&participants, "SELECT DISTINCT UserId FROM Posts WHERE RootId=:RootId OR Id=:RootId", map[string]interface{}{"RootId": rootId}); err != nil {
				return err
			}
			// calculate reply count
			count, err := transaction.SelectInt("SELECT COUNT(Id) FROM Posts WHERE RootId=:RootId", map[string]interface{}{"RootId": rootId})
			if err != nil {
				return err
			}
			// no metadata entry, create one
			if err := transaction.Insert(&model.Thread{
				PostId:       rootId,
				ChannelId:    posts[0].ChannelId,
				ReplyCount:   count,
				LastReplyAt:  now,
				Participants: participants,
			}); err != nil {
				return err
			}
		} else {
			// metadata exists, update it
			thread.LastReplyAt = now
			for _, post := range posts {
				thread.ReplyCount += 1
				if !thread.Participants.Contains(post.UserId) {
					thread.Participants = append(thread.Participants, post.UserId)
				}
			}
			if _, err := transaction.Update(thread); err != nil {
				return err
			}
		}
	}
	return nil
}
