package sqlstore

import (
	"sync/atomic"
	"github.com/go-gorp/gorp"
	"github.com/masterhung0112/go_server/model"
	"github.com/masterhung0112/go_server/store"

)

type SqlSupplierStores struct {
  user                 store.UserStore
}

type SqlSupplier struct {
  // rrCounter and srCounter should be kept first.
	// See https://github.com/mattermost/mattermost-server/v5/pull/7281
	rrCounter      int64
  srCounter      int64

  master         *gorp.DbMap
  replicas       []*gorp.DbMap
  stores         SqlSupplierStores
  settings       *model.SqlSettings
  lockedToMaster bool
  license        *model.License
}

func (ss *SqlSupplier) DriverName() string {
	return *ss.settings.DriverName
}

func (ss *SqlSupplier) GetMaster() *gorp.DbMap {
	return ss.master
}

func (ss *SqlSupplier) GetReplica() *gorp.DbMap {
	if len(ss.settings.DataSourceReplicas) == 0 || ss.lockedToMaster || ss.license == nil {
		return ss.GetMaster()
	}

	rrNum := atomic.AddInt64(&ss.rrCounter, 1) % int64(len(ss.replicas))
	return ss.replicas[rrNum]
}

func (ss *SqlSupplier) User() store.UserStore {
	return ss.stores.user
}