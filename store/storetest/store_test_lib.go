package storetest

import (
	"github.com/mattermost/gorp"
)

type SqlStore interface {
	GetMaster() *gorp.DbMap
	DriverName() string
}
