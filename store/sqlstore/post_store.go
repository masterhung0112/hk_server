package sqlstore

import (
	"github.com/masterhung0112/hk_server/einterfaces"
	"github.com/masterhung0112/hk_server/model"
	"github.com/masterhung0112/hk_server/store"
	"sync"
)

type SqlPostStore struct {
	*SqlSupplier
	metrics           einterfaces.MetricsInterface
	maxPostSizeOnce   sync.Once
	maxPostSizeCached int
}

func (s *SqlPostStore) ClearCaches() {
}

func postSliceColumns() []string {
	return []string{"Id", "CreateAt", "UpdateAt", "EditAt", "DeleteAt", "IsPinned", "UserId", "ChannelId", "RootId", "ParentId", "OriginalId", "Message", "Type", "Props", "Hashtags", "Filenames", "FileIds", "HasReactions"}
}

func postToSlice(post *model.Post) []interface{} {
	return []interface{}{
		post.Id,
		post.CreateAt,
		post.UpdateAt,
		post.EditAt,
		post.DeleteAt,
		post.IsPinned,
		post.UserId,
		post.ChannelId,
		post.RootId,
		post.ParentId,
		post.OriginalId,
		post.Message,
		post.Type,
		model.StringInterfaceToJson(post.Props),
		post.Hashtags,
		model.ArrayToJson(post.Filenames),
		model.ArrayToJson(post.FileIds),
		post.HasReactions,
	}
}

func newSqlPostStore(sqlSupplier *SqlSupplier, metrics einterfaces.MetricsInterface) store.PostStore {
	s := &SqlPostStore{
		SqlSupplier:       sqlSupplier,
		metrics:           metrics,
		maxPostSizeCached: model.POST_MESSAGE_MAX_RUNES_V1,
	}

	for _, db := range sqlSupplier.GetAllConns() {
		table := db.AddTableWithName(model.Post{}, "Posts").SetKeys(false, "Id")
		table.ColMap("Id").SetMaxSize(26)
		table.ColMap("UserId").SetMaxSize(26)
		table.ColMap("ChannelId").SetMaxSize(26)
		table.ColMap("RootId").SetMaxSize(26)
		table.ColMap("ParentId").SetMaxSize(26)
		table.ColMap("OriginalId").SetMaxSize(26)
		table.ColMap("Message").SetMaxSize(model.POST_MESSAGE_MAX_BYTES_V2)
		table.ColMap("Type").SetMaxSize(26)
		table.ColMap("Hashtags").SetMaxSize(1000)
		table.ColMap("Props").SetMaxSize(8000)
		table.ColMap("Filenames").SetMaxSize(model.POST_FILENAMES_MAX_RUNES)
		table.ColMap("FileIds").SetMaxSize(150)
	}

	return s
}
