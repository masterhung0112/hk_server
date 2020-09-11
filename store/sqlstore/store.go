package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/go_server/store"
	"github.com/mattermost/gorp"
)

type SqlStore interface {
	DriverName() string
	GetMaster() *gorp.DbMap
	GetReplica() *gorp.DbMap

	User() store.UserStore
	Team() store.TeamStore
	Role() store.RoleStore
	Scheme() store.SchemeStore
	Channel() store.ChannelStore

	GetAllConns() []*gorp.DbMap
	getQueryBuilder() sq.StatementBuilderType
}
