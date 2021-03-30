// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/masterhung0112/hk_server/v5/store"
	"github.com/masterhung0112/hk_server/v5/store/sqlstore"
	"github.com/masterhung0112/hk_server/v5/store/storetest"
)

func TestStoreUpgrade(t *testing.T) {
	storetest.StoreTest(t, func(t *testing.T, ss store.Store) {
		sqlStore := ss.(*sqlstore.SqlStore)

		t.Run("invalid currentModelVersion", func(t *testing.T) {
			err := sqlstore.UpgradeDatabase(sqlStore, "notaversion")
			require.EqualError(t, err, "failed to parse current model version notaversion: No Major.Minor.Patch elements found")
		})

		t.Run("upgrade from invalid version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "invalid")
			err := sqlstore.UpgradeDatabase(sqlStore, "5.8.0")
			require.EqualError(t, err, "failed to parse database schema version invalid: No Major.Minor.Patch elements found")
			require.Equal(t, "invalid", sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade from unsupported version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "2.0.0")
			err := sqlstore.UpgradeDatabase(sqlStore, "5.8.0")
			require.EqualError(t, err, "Database schema version 2.0.0 is no longer supported. This Mattermost server supports automatic upgrades from schema version 3.0.0 through schema version 5.8.0. Please manually upgrade to at least version 3.0.0 before continuing.")
			require.Equal(t, "2.0.0", sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade from earliest supported version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, sqlstore.Version300)
			err := sqlstore.UpgradeDatabase(sqlStore, sqlstore.CurrentSchemaVersion)
			require.NoError(t, err)
			require.Equal(t, sqlstore.CurrentSchemaVersion, sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade from no existing version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "")
			err := sqlstore.UpgradeDatabase(sqlStore, sqlstore.CurrentSchemaVersion)
			require.NoError(t, err)
			require.Equal(t, sqlstore.CurrentSchemaVersion, sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade schema running earlier minor version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "5.1.0")
			err := sqlstore.UpgradeDatabase(sqlStore, "5.8.0")
			require.NoError(t, err)
			// Assert sqlstore.CurrentSchemaVersion, not 5.8.0, since the migrations will move
			// past 5.8.0 regardless of the input parameter.
			require.Equal(t, sqlstore.CurrentSchemaVersion, sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade schema running later minor version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "5.99.0")
			err := sqlstore.UpgradeDatabase(sqlStore, "5.8.0")
			require.NoError(t, err)
			require.Equal(t, "5.99.0", sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade schema running earlier major version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "4.1.0")
			err := sqlstore.UpgradeDatabase(sqlStore, sqlstore.CurrentSchemaVersion)
			require.NoError(t, err)
			require.Equal(t, sqlstore.CurrentSchemaVersion, sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("upgrade schema running later major version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, "6.0.0")
			err := sqlstore.UpgradeDatabase(sqlStore, "5.8.0")
			require.EqualError(t, err, "Database schema version 6.0.0 is not supported. This Mattermost server supports only >=5.8.0, <6.0.0. Please upgrade to at least version 6.0.0 before continuing.")
			require.Equal(t, "6.0.0", sqlStore.GetCurrentSchemaVersion())
		})
	})
}

func TestSaveSchemaVersion(t *testing.T) {
	storetest.StoreTest(t, func(t *testing.T, ss store.Store) {
		sqlStore := ss.(*sqlstore.SqlStore)

		t.Run("set earliest version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, sqlstore.Version300)
			props, err := ss.System().Get()
			require.NoError(t, err)

			require.Equal(t, sqlstore.Version300, props["Version"])
			require.Equal(t, sqlstore.Version300, sqlStore.GetCurrentSchemaVersion())
		})

		t.Run("set current version", func(t *testing.T) {
			sqlstore.SaveSchemaVersion(sqlStore, sqlstore.CurrentSchemaVersion)
			props, err := ss.System().Get()
			require.NoError(t, err)

			require.Equal(t, sqlstore.CurrentSchemaVersion, props["Version"])
			require.Equal(t, sqlstore.CurrentSchemaVersion, sqlStore.GetCurrentSchemaVersion())
		})
	})
}
