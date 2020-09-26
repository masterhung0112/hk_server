package sqlstore

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/mattermost/gorp"
)

type SqlPreferenceStore struct {
	SqlStore
}

func newSqlPreferenceStore(sqlStore SqlStore) store.PreferenceStore {
	s := &SqlPreferenceStore{sqlStore}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(model.Preference{}, "Preferences").SetKeys(false, "UserId", "Category", "Name")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Category").SetMaxSize(32)
		table.ColMap("Name").SetMaxSize(32)
		table.ColMap("Value").SetMaxSize(2000)
	}

	return s
}

func (s SqlPreferenceStore) Save(preferences *model.Preferences) error {
	// wrap in a transaction so that if one fails, everything fails
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return errors.Wrap(err, "begin_transaction")
	}

	defer finalizeTransaction(transaction)
	for _, preference := range *preferences {
		preference := preference
		if upsertErr := s.save(transaction, &preference); upsertErr != nil {
			return upsertErr
		}
	}

	if err := transaction.Commit(); err != nil {
		// don't need to rollback here since the transaction is already closed
		return errors.Wrap(err, "commit_transaction")
	}
	return nil
}

func (s SqlPreferenceStore) save(transaction *gorp.Transaction, preference *model.Preference) error {
	preference.PreUpdate()

	if err := preference.IsValid(); err != nil {
		return err
	}

	params := map[string]interface{}{
		"UserId":   preference.UserId,
		"Category": preference.Category,
		"Name":     preference.Name,
		"Value":    preference.Value,
	}

	if s.DriverName() == model.DATABASE_DRIVER_MYSQL {
		if _, err := transaction.Exec(
			`INSERT INTO
				Preferences
				(UserId, Category, Name, Value)
			VALUES
				(:UserId, :Category, :Name, :Value)
			ON DUPLICATE KEY UPDATE
				Value = :Value`, params); err != nil {
			return errors.Wrap(err, "failed to save Preference")
		}
		return nil
	} else if s.DriverName() == model.DATABASE_DRIVER_POSTGRES {
		// postgres has no way to upsert values until version 9.5 and trying inserting and then updating causes transactions to abort
		count, err := transaction.SelectInt(
			`SELECT
				count(0)
			FROM
				Preferences
			WHERE
				UserId = :UserId
				AND Category = :Category
				AND Name = :Name`, params)
		if err != nil {
			return errors.Wrap(err, "failed to count Preferences")
		}

		if count == 1 {
			return s.update(transaction, preference)
		}
		return s.insert(transaction, preference)
	}
	return store.NewErrNotImplemented("failed to update preference because of missing driver")
}

func (s SqlPreferenceStore) Get(userId string, category string, name string) (*model.Preference, error) {
	var preference *model.Preference

	if err := s.GetReplica().SelectOne(&preference,
		`SELECT
			*
		FROM
			Preferences
		WHERE
			UserId = :UserId
			AND Category = :Category
			AND Name = :Name`, map[string]interface{}{"UserId": userId, "Category": category, "Name": name}); err != nil {
		return nil, errors.Wrapf(err, "failed to find Preference with userId=%s, category=%s, name=%s", userId, category, name)
	}
	return preference, nil
}

func (s SqlPreferenceStore) insert(transaction *gorp.Transaction, preference *model.Preference) error {
	if err := transaction.Insert(preference); err != nil {
		if IsUniqueConstraintError(err, []string{"UserId", "preferences_pkey"}) {
			return store.NewErrInvalidInput("Preference", "<userId, category, name>", fmt.Sprintf("<%s, %s, %s>", preference.UserId, preference.Category, preference.Name))
		}
		return errors.Wrapf(err, "failed to save Preference with userId=%s, category=%s, name=%s", preference.UserId, preference.Category, preference.Name)
	}

	return nil
}

func (s SqlPreferenceStore) update(transaction *gorp.Transaction, preference *model.Preference) error {
	if _, err := transaction.Update(preference); err != nil {
		return errors.Wrapf(err, "failed to update Preference with userId=%s, category=%s, name=%s", preference.UserId, preference.Category, preference.Name)
	}

	return nil
}
