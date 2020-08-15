package testlib

import (
	"github.com/masterhung0112/go_server/store"
)

type TestStore struct {
	store.Store
}

func (s *TestStore) Close() {
	// Don't propagate to the underlying store, since this instance is persistent.
}