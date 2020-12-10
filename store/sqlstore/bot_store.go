package sqlstore

import (
	"database/sql"
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"github.com/pkg/errors"
)

// bot is a subset of the model.Bot type, omitting the model.User fields.
type bot struct {
	UserId         string `json:"user_id"`
	Description    string `json:"description"`
	OwnerId        string `json:"owner_id"`
	LastIconUpdate int64  `json:"last_icon_update"`
	CreateAt       int64  `json:"create_at"`
	UpdateAt       int64  `json:"update_at"`
	DeleteAt       int64  `json:"delete_at"`
}

func botFromModel(b *model.Bot) *bot {
	return &bot{
		UserId:         b.UserId,
		Description:    b.Description,
		OwnerId:        b.OwnerId,
		LastIconUpdate: b.LastIconUpdate,
		CreateAt:       b.CreateAt,
		UpdateAt:       b.UpdateAt,
		DeleteAt:       b.DeleteAt,
	}
}

// SqlBotStore is a store for managing bots in the database.
// Bots are otherwise normal users with extra metadata record in the Bots table. The primary key
// for a bot matches the primary key value for corresponding User record.
type SqlBotStore struct {
	*SqlStore
	metrics einterfaces.MetricsInterface
}

// newSqlBotStore creates an instance of SqlBotStore, registering the table schema in question.
func newSqlBotStore(sqlStore *SqlStore, metrics einterfaces.MetricsInterface) store.BotStore {
	us := &SqlBotStore{
		SqlStore: sqlStore,
		metrics:  metrics,
	}

	for _, db := range sqlStore.GetAllConns() {
		table := db.AddTableWithName(bot{}, "Bots").SetKeys(false, "UserId")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Description").SetMaxSize(1024)
		table.ColMap("OwnerId").SetMaxSize(model.BOT_CREATOR_ID_MAX_RUNES)
	}

	return us
}

func (us SqlBotStore) Get(botUserId string, includeDeleted bool) (*model.Bot, error) {
	var excludeDeletedSql = "AND b.DeleteAt = 0"
	if includeDeleted {
		excludeDeletedSql = ""
	}

	query := `
		SELECT
			b.UserId,
			u.Username,
			u.FirstName AS DisplayName,
			b.Description,
			b.OwnerId,
			COALESCE(b.LastIconUpdate, 0) AS LastIconUpdate,
			b.CreateAt,
			b.UpdateAt,
			b.DeleteAt
		FROM
			Bots b
		JOIN
			Users u ON (u.Id = b.UserId)
		WHERE
			b.UserId = :user_id
			` + excludeDeletedSql + `
	`

	var bot *model.Bot
	if err := us.GetReplica().SelectOne(&bot, query, map[string]interface{}{"user_id": botUserId}); err == sql.ErrNoRows {
		return nil, store.NewErrNotFound("Bot", botUserId)
	} else if err != nil {
		return nil, errors.Wrapf(err, "selectone: user_id=%s", botUserId)
	}

	return bot, nil
}

// Save persists a new bot to the database.
// It assumes the corresponding user was saved via the user store.
func (us SqlBotStore) Save(bot *model.Bot) (*model.Bot, error) {
	bot = bot.Clone()
	bot.PreSave()

	if err := bot.IsValid(); err != nil { // TODO: change to return error in v6.
		return nil, err
	}

	if err := us.GetMaster().Insert(botFromModel(bot)); err != nil {
		return nil, errors.Wrapf(err, "insert: user_id=%s", bot.UserId)
	}

	return bot, nil
}

// PermanentDelete removes the bot from the database altogether.
// If the corresponding user is to be deleted, it must be done via the user store.
func (us SqlBotStore) PermanentDelete(botUserId string) error {
	query := "DELETE FROM Bots WHERE UserId = :user_id"
	if _, err := us.GetMaster().Exec(query, map[string]interface{}{"user_id": botUserId}); err != nil {
		return store.NewErrInvalidInput("Bot", "UserId", botUserId)
	}
	return nil
}
