package storetest

import (
	"testing"

	"github.com/masterhung0112/go_server/store"
	"github.com/stretchr/testify/require"
)

func TestUserStore(t *testing.T, ss store.Store, s SqlSupplier) {
	users, err := ss.User().GetAll()
	require.Nil(t, err, "failed cleaning up test users")

	for _, u := range users {
		err := ss.User().PermanentDelete(u.Id)
		require.Nil(t, err, "failed cleaning up test user %s", u.Username)
	}
}
