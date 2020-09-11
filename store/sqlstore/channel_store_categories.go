package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/go_server/model"
	"github.com/mattermost/gorp"
	"net/http"
)

type sidebarCategoryForJoin struct {
	model.SidebarCategory
	ChannelId *string
}

func (s SqlChannelStore) completePopulatingCategoryChannels(category *model.SidebarCategoryWithChannels) (*model.SidebarCategoryWithChannels, *model.AppError) {
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return nil, model.NewAppError("SqlChannelStore.completePopulatingCategoryChannels", "store.sql_channel.sidebar_categories.open_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}
	defer finalizeTransaction(transaction)

	result, appErr := s.completePopulatingCategoryChannelsT(transaction, category)
	if appErr != nil {
		return nil, appErr
	}

	if err = transaction.Commit(); err != nil {
		return nil, model.NewAppError("SqlChannelStore.completePopulatingCategoryChannels", "store.sql_channel.sidebar_categories.commit_transaction.app_error", nil, err.Error(), http.StatusInternalServerError)
	}

	return result, nil
}

func (s SqlChannelStore) completePopulatingCategoryChannelsT(transation *gorp.Transaction, category *model.SidebarCategoryWithChannels) (*model.SidebarCategoryWithChannels, *model.AppError) {
	if category.Type == model.SidebarCategoryCustom || category.Type == model.SidebarCategoryFavorites {
		return category, nil
	}

	var channelTypeFilter sq.Sqlizer
	if category.Type == model.SidebarCategoryDirectMessages {
		// any DM/GM channels that aren't in any category should be returned as part of the Direct Messages category
		channelTypeFilter = sq.Eq{"Channels.Type": []string{model.CHANNEL_DIRECT, model.CHANNEL_GROUP}}
	} else if category.Type == model.SidebarCategoryChannels {
		// any public/private channels that are on the current team and aren't in any category should be returned as part of the Channels category
		channelTypeFilter = sq.And{
			sq.Eq{"Channels.Type": []string{model.CHANNEL_OPEN, model.CHANNEL_PRIVATE}},
			sq.Eq{"Channels.TeamId": category.TeamId},
		}
	}

	// A subquery that is true if the channel does not have a SidebarChannel entry for the current user on the current team
	doesNotHaveSidebarChannel := sq.Select("1").
		Prefix("NOT EXISTS (").
		From("SidebarChannels").
		Join("SidebarCategories on SidebarChannels.CategoryId=SidebarCategories.Id").
		Where(sq.And{
			sq.Expr("SidebarChannels.ChannelId = ChannelMembers.ChannelId"),
			sq.Eq{"SidebarCategories.UserId": category.UserId},
			sq.Eq{"SidebarCategories.TeamId": category.TeamId},
		}).
		Suffix(")")

	var channels []string
	sql, args, _ := s.getQueryBuilder().
		Select("Id").
		From("ChannelMembers").
		LeftJoin("Channels ON Channels.Id=ChannelMembers.ChannelId").
		Where(sq.And{
			sq.Eq{"ChannelMembers.UserId": category.UserId},
			channelTypeFilter,
			sq.Eq{"Channels.DeleteAt": 0},
			doesNotHaveSidebarChannel,
		}).
		OrderBy("DisplayName ASC").ToSql()

	if _, err := transation.Select(&channels, sql, args...); err != nil {
		return nil, model.NewAppError("SqlPostStore.completePopulatingCategoryChannelsT", "store.sql_channel.sidebar_categories.app_error", nil, err.Error(), http.StatusNotFound)
	}

	category.Channels = append(channels, category.Channels...)
	return category, nil
}

func (s SqlChannelStore) GetSidebarCategory(categoryId string) (*model.SidebarCategoryWithChannels, *model.AppError) {
	var categories []*sidebarCategoryForJoin
	sql, args, _ := s.getQueryBuilder().
		Select("SidebarCategories.*", "SidebarChannels.ChannelId").
		From("SidebarCategories").
		LeftJoin("SidebarChannels ON SidebarChannels.CategoryId=SidebarCategories.Id").
		Where(sq.Eq{"SidebarCategories.Id": categoryId}).
		OrderBy("SidebarChannels.SortOrder ASC").ToSql()
	if _, err := s.GetReplica().Select(&categories, sql, args...); err != nil {
		return nil, model.NewAppError("SqlPostStore.GetSidebarCategory", "store.sql_channel.sidebar_categories.app_error", nil, err.Error(), http.StatusNotFound)
	}
	result := &model.SidebarCategoryWithChannels{
		SidebarCategory: categories[0].SidebarCategory,
		Channels:        make([]string, 0),
	}
	for _, category := range categories {
		if category.ChannelId != nil {
			result.Channels = append(result.Channels, *category.ChannelId)
		}
	}
	return s.completePopulatingCategoryChannels(result)
}
