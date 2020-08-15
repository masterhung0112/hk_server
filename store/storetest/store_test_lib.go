package storetest

import (
  "github.com/mattermost/gorp"
)

type SqlSupplier interface {
	GetMaster() *gorp.DbMap
	DriverName() string
}