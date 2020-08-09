package storetest

import (
  "github.com/go-gorp/gorp"
)

type SqlSupplier interface {
	GetMaster() *gorp.DbMap
	DriverName() string
}