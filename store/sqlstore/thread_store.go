package sqlstore

import (
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

type SqlThreadStore struct {
	*SqlStore
}

func (s *SqlThreadStore) ClearCaches() {
}

func newSqlThreadStore(sqlStore *SqlStore) store.ThreadStore {
	s := &SqlThreadStore{
		SqlStore: sqlStore,
	}

	for _, db := range sqlStore.GetAllConns() {
		tableThreads := db.AddTableWithName(model.Thread{}, "Threads").SetKeys(false, "PostId")
		tableThreads.ColMap("PostId").SetMaxSize(26)
		tableThreads.ColMap("ChannelId").SetMaxSize(26)
		tableThreads.ColMap("Participants").SetMaxSize(0)
		tableThreadMemberships := db.AddTableWithName(model.ThreadMembership{}, "ThreadMemberships").SetKeys(false, "PostId", "UserId")
		tableThreadMemberships.ColMap("PostId").SetMaxSize(26)
		tableThreadMemberships.ColMap("UserId").SetMaxSize(26)
	}

	return s
}

func (s *SqlThreadStore) Update(thread *model.Thread) (*model.Thread, error) {
	if _, err := s.GetMaster().Update(thread); err != nil {
		return nil, errors.Wrapf(err, "failed to update thread with id=%s", thread.PostId)
	}

	return thread, nil
}

func (s *SqlThreadStore) Get(id string) (*model.Thread, error) {
	var thread model.Thread
	query, args, _ := s.getQueryBuilder().Select("*").From("Threads").Where(sq.Eq{"PostId": id}).ToSql()
	err := s.GetReplica().SelectOne(&thread, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, store.NewErrNotFound("Thread", id)
		}

		return nil, errors.Wrapf(err, "failed to get thread with id=%s", id)
	}
	return &thread, nil
}

func threadSliceColumns() []string {
	return []string{"PostId", "ChannelId", "LastReplyAt", "ReplyCount", "Participants"}
}

func (s *SqlThreadStore) CollectThreadsWithNewerReplies(userId string, channelIds []string, timestamp int64) ([]string, error) {
	var changedThreads []string
	query, args, _ := s.getQueryBuilder().
		Select("Threads.PostId").
		From("Threads").
		LeftJoin("ChannelMembers ON ChannelMembers.ChannelId=Threads.ChannelId").
		Where(sq.And{
			sq.Eq{"Threads.ChannelId": channelIds},
			sq.Eq{"ChannelMembers.UserId": userId},
			sq.Or{
				sq.Expr("Threads.LastReplyAt >= ChannelMembers.LastViewedAt"),
				sq.GtOrEq{"Threads.LastReplyAt": timestamp},
			},
		}).
		ToSql()
	if _, err := s.GetReplica().Select(&changedThreads, query, args...); err != nil {
		return nil, errors.Wrap(err, "failed to fetch threads")
	}
	return changedThreads, nil
}

func (s *SqlThreadStore) UpdateUnreadsByChannel(userId string, changedThreads []string, timestamp int64, updateViewedTimestamp bool) error {
	if len(changedThreads) == 0 {
		return nil
	}

	qb := s.getQueryBuilder().
		Update("ThreadMemberships").
		Where(sq.Eq{"UserId": userId, "PostId": changedThreads}).
		Set("LastUpdated", timestamp)

	if updateViewedTimestamp {
		qb = qb.Set("LastViewed", timestamp)
	}
	updateQuery, updateArgs, _ := qb.ToSql()

	if _, err := s.GetMaster().Exec(updateQuery, updateArgs...); err != nil {
		return errors.Wrap(err, "failed to update thread membership")
	}

	return nil
}
