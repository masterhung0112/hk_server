package sqlstore

import (
	"fmt"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"

	"github.com/pkg/errors"
)

const (
	SESSIONS_CLEANUP_DELAY_MILLISECONDS = 100
)

type SqlSessionStore struct {
	SqlStore
}

func newSqlSessionStore(sqlStore SqlStore) store.SessionStore {
	us := &SqlSessionStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Session{}, "Sessions").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("Token").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("DeviceId").SetMaxSize(512)
		table.ColMap("Roles").SetMaxSize(64)
		table.ColMap("Props").SetMaxSize(1000)
	}

	return us
}

func (me SqlSessionStore) Save(session *model.Session) (*model.Session, error) {
	if len(session.Id) > 0 {
		return nil, store.NewErrInvalidInput("Session", "id", session.Id)
	}
	session.PreSave()

	if err := me.GetMaster().Insert(session); err != nil {
		return nil, errors.Wrapf(err, "failed to save Session with id=%s", session.Id)
	}

	teamMembers, err := me.Team().GetTeamsForUser(session.UserId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find TeamMembers for Session with userId=%s", session.UserId)
	}

	session.TeamMembers = make([]*model.TeamMember, 0, len(teamMembers))
	for _, tm := range teamMembers {
		if tm.DeleteAt == 0 {
			session.TeamMembers = append(session.TeamMembers, tm)
		}
	}

	return session, nil
}

func (me SqlSessionStore) Get(sessionIdOrToken string) (*model.Session, error) {
	var sessions []*model.Session

	if _, err := me.GetReplica().Select(&sessions, "SELECT * FROM Sessions WHERE Token = :Token OR Id = :Id LIMIT 1", map[string]interface{}{"Token": sessionIdOrToken, "Id": sessionIdOrToken}); err != nil {
		return nil, errors.Wrapf(err, "failed to find Sessions with sessionIdOrToken=%s", sessionIdOrToken)
	} else if len(sessions) == 0 {
		return nil, store.NewErrNotFound("Session", fmt.Sprintf("sessionIdOrToken=%s", sessionIdOrToken))
	}
	session := sessions[0]

	tempMembers, err := me.Team().GetTeamsForUser(session.UserId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to find TeamMembers for Session with userId=%s", session.UserId)
	}
	sessions[0].TeamMembers = make([]*model.TeamMember, 0, len(tempMembers))
	for _, tm := range tempMembers {
		if tm.DeleteAt == 0 {
			sessions[0].TeamMembers = append(sessions[0].TeamMembers, tm)
		}
	}
	return session, nil
}

func (me SqlSessionStore) Remove(sessionIdOrToken string) error {
	_, err := me.GetMaster().Exec("DELETE FROM Sessions WHERE Id = :Id Or Token = :Token", map[string]interface{}{"Id": sessionIdOrToken, "Token": sessionIdOrToken})
	if err != nil {
		return errors.Wrapf(err, "failed to delete Session with sessionIdOrToken=%s", sessionIdOrToken)
	}
	return nil
}
