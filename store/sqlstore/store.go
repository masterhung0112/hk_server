package sqlstore

import (
	"github.com/mattermost/gorp"
	"github.com/masterhung0112/go_server/store"
)

type SqlStore interface {
  GetMaster() *gorp.DbMap
  GetReplica() *gorp.DbMap

  User() store.UserStore
}