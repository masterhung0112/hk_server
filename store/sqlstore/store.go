package sqlstore

import (
	"github.com/mattermost/gorp"
  "github.com/masterhung0112/go_server/store"
  sq "github.com/Masterminds/squirrel"
)

type SqlStore interface {
  GetMaster() *gorp.DbMap
  GetReplica() *gorp.DbMap

  User() store.UserStore

  GetAllConns() []*gorp.DbMap
  getQueryBuilder() sq.StatementBuilderType
}