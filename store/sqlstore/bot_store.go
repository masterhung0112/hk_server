package sqlstore

import (
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
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
	*SqlSupplier
	metrics einterfaces.MetricsInterface
}

// newSqlBotStore creates an instance of SqlBotStore, registering the table schema in question.
func newSqlBotStore(sqlSupplier *SqlSupplier, metrics einterfaces.MetricsInterface) store.BotStore {
	us := &SqlBotStore{
		SqlSupplier: sqlSupplier,
		metrics:     metrics,
	}

	for _, db := range sqlSupplier.GetAllConns() {
		table := db.AddTableWithName(bot{}, "Bots").SetKeys(false, "UserId")
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("Description").SetMaxSize(1024)
		table.ColMap("OwnerId").SetMaxSize(model.BOT_CREATOR_ID_MAX_RUNES)
	}

	return us
}
