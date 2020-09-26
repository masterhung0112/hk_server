package sqlstore

import (
	sq "github.com/Masterminds/squirrel"
	"github.com/masterhung0112/hk_server/model"
	"github.com/mattermost/gorp"
	"github.com/pkg/errors"
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

func (s SqlChannelStore) CreateInitialSidebarCategories(userId, teamId string) error {
	transaction, err := s.GetMaster().Begin()
	if err != nil {
		return errors.Wrap(err, "CreateInitialSidebarCategories: begin_transaction")
	}
	defer finalizeTransaction(transaction)

	if err := s.createInitialSidebarCategoriesT(transaction, userId, teamId); err != nil {
		return errors.Wrap(err, "CreateInitialSidebarCategories: createInitialSidebarCategoriesT")
	}

	if err := transaction.Commit(); err != nil {
		return errors.Wrap(err, "CreateInitialSidebarCategories: commit_transaction")
	}

	return nil
}

func (s SqlChannelStore) createInitialSidebarCategoriesT(transaction *gorp.Transaction, userId, teamId string) error {
	selectQuery, selectParams, _ := s.getQueryBuilder().
		Select("Type").
		From("SidebarCategories").
		Where(sq.Eq{
			"UserId": userId,
			"TeamId": teamId,
			"Type":   []model.SidebarCategoryType{model.SidebarCategoryFavorites, model.SidebarCategoryChannels, model.SidebarCategoryDirectMessages},
		}).ToSql()

	var existingTypes []model.SidebarCategoryType
	_, err := transaction.Select(&existingTypes, selectQuery, selectParams...)
	if err != nil {
		return errors.Wrap(err, "createInitialSidebarCategoriesT: failed to select existing categories")
	}

	hasCategoryOfType := make(map[model.SidebarCategoryType]bool, len(existingTypes))
	for _, existingType := range existingTypes {
		hasCategoryOfType[existingType] = true
	}

	if !hasCategoryOfType[model.SidebarCategoryFavorites] {
		favoritesCategoryId := model.NewId()

		// Create the SidebarChannels first since there's more opportunity for something to fail here
		if err := s.migrateFavoritesToSidebarT(transaction, userId, teamId, favoritesCategoryId); err != nil {
			return errors.Wrap(err, "createInitialSidebarCategoriesT: failed to migrate favorites to sidebar")
		}

		if err := transaction.Insert(&model.SidebarCategory{
			DisplayName: "Favorites", // This will be retranslated by the client into the user's locale
			Id:          favoritesCategoryId,
			UserId:      userId,
			TeamId:      teamId,
			Sorting:     model.SidebarCategorySortDefault,
			SortOrder:   model.DefaultSidebarSortOrderFavorites,
			Type:        model.SidebarCategoryFavorites,
		}); err != nil {
			return errors.Wrap(err, "createInitialSidebarCategoriesT: failed to insert favorites category")
		}
	}

	if !hasCategoryOfType[model.SidebarCategoryChannels] {
		if err := transaction.Insert(&model.SidebarCategory{
			DisplayName: "Channels", // This will be retranslateed by the client into the user's locale
			Id:          model.NewId(),
			UserId:      userId,
			TeamId:      teamId,
			Sorting:     model.SidebarCategorySortDefault,
			SortOrder:   model.DefaultSidebarSortOrderChannels,
			Type:        model.SidebarCategoryChannels,
		}); err != nil {
			return errors.Wrap(err, "createInitialSidebarCategoriesT: failed to insert channels category")
		}
	}

	if !hasCategoryOfType[model.SidebarCategoryDirectMessages] {
		if err := transaction.Insert(&model.SidebarCategory{
			DisplayName: "Direct Messages", // This will be retranslateed by the client into the user's locale
			Id:          model.NewId(),
			UserId:      userId,
			TeamId:      teamId,
			Sorting:     model.SidebarCategorySortRecent,
			SortOrder:   model.DefaultSidebarSortOrderDMs,
			Type:        model.SidebarCategoryDirectMessages,
		}); err != nil {
			return errors.Wrap(err, "createInitialSidebarCategoriesT: failed to insert direct messages category")
		}
	}

	return nil
}

func (s SqlChannelStore) migrateFavoritesToSidebarT(transaction *gorp.Transaction, userId, teamId, favoritesCategoryId string) error {
	favoritesQuery, favoritesParams, _ := s.getQueryBuilder().
		Select("Preferences.Name").
		From("Preferences").
		Join("Channels on Preferences.Name = Channels.Id").
		Join("ChannelMembers on Preferences.Name = ChannelMembers.ChannelId and Preferences.UserId = ChannelMembers.UserId").
		Where(sq.Eq{
			"Preferences.UserId":   userId,
			"Preferences.Category": model.PREFERENCE_CATEGORY_FAVORITE_CHANNEL,
			"Preferences.Value":    "true",
		}).
		Where(sq.Or{
			sq.Eq{"Channels.TeamId": teamId},
			sq.Eq{"Channels.TeamId": ""},
		}).
		OrderBy(
			"Channels.DisplayName",
			"Channels.Name ASC",
		).ToSql()

	var favoriteChannelIds []string
	if _, err := transaction.Select(&favoriteChannelIds, favoritesQuery, favoritesParams...); err != nil {
		return errors.Wrap(err, "migrateFavoritesToSidebarT: unable to get favorite channel IDs")
	}

	for i, channelId := range favoriteChannelIds {
		if err := transaction.Insert(&model.SidebarChannel{
			ChannelId:  channelId,
			CategoryId: favoritesCategoryId,
			UserId:     userId,
			SortOrder:  int64(i * model.MinimalSidebarSortDistance),
		}); err != nil {
			return errors.Wrap(err, "migrateFavoritesToSidebarT: unable to insert SidebarChannel")
		}
	}

	return nil
}
