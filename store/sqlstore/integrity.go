// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package sqlstore

import (
	"github.com/masterhung0112/hk_server/v5/model"
	"github.com/masterhung0112/hk_server/v5/shared/mlog"

	sq "github.com/Masterminds/squirrel"
)

type RelationalCheckConfig struct {
	ParentName         string
	ParentIdAttr       string
	ChildName          string
	ChildIdAttr        string
	CanParentIdBeEmpty bool
	SortRecords        bool
	Filter             interface{}
}

func getOrphanedRecords(ss *SqlStore, cfg RelationalCheckConfig) ([]model.OrphanedRecord, error) {
	var records []model.OrphanedRecord

	sub := ss.getQueryBuilder().
		Select("TRUE").
		From(cfg.ParentName + " AS PT").
		Prefix("NOT EXISTS (").
		Suffix(")").
		Where("PT.id = CT." + cfg.ParentIdAttr)

	main := ss.getQueryBuilder().
		Select().
		Column("CT." + cfg.ParentIdAttr + " AS ParentId").
		From(cfg.ChildName + " AS CT").
		Where(sub)

	if cfg.ChildIdAttr != "" {
		main = main.Column("CT." + cfg.ChildIdAttr + " AS ChildId")
	}

	if cfg.CanParentIdBeEmpty {
		main = main.Where(sq.NotEq{"CT." + cfg.ParentIdAttr: ""})
	}

	if cfg.Filter != nil {
		main = main.Where(cfg.Filter)
	}

	if cfg.SortRecords {
		main = main.OrderBy("CT." + cfg.ParentIdAttr)
	}

	query, args, _ := main.ToSql()

	_, err := ss.GetMaster().Select(&records, query, args...)

	return records, err
}

func checkParentChildIntegrity(ss *SqlStore, config RelationalCheckConfig) model.IntegrityCheckResult {
	var result model.IntegrityCheckResult
	var data model.RelationalIntegrityCheckData

	config.SortRecords = true
	data.Records, result.Err = getOrphanedRecords(ss, config)
	if result.Err != nil {
		mlog.Error("Error while getting orphaned records", mlog.Err(result.Err))
		return result
	}
	data.ParentName = config.ParentName
	data.ChildName = config.ChildName
	data.ParentIdAttr = config.ParentIdAttr
	data.ChildIdAttr = config.ChildIdAttr
	result.Data = data

	return result
}

func checkChannelsCommandWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "CommandWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkChannelsChannelMemberHistoryIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "ChannelMemberHistory",
		ChildIdAttr:  "",
	})
}

func checkChannelsChannelMembersIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "ChannelMembers",
		ChildIdAttr:  "",
	})
}

func checkChannelsIncomingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "IncomingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkChannelsOutgoingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "OutgoingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkChannelsPostsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Channels",
		ParentIdAttr: "ChannelId",
		ChildName:    "Posts",
		ChildIdAttr:  "Id",
	})
}

func checkCommandsCommandWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Commands",
		ParentIdAttr: "CommandId",
		ChildName:    "CommandWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkPostsFileInfoIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Posts",
		ParentIdAttr: "PostId",
		ChildName:    "FileInfo",
		ChildIdAttr:  "Id",
	})
}

func checkPostsPostsParentIdIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Posts",
		ParentIdAttr:       "ParentId",
		ChildName:          "Posts",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkPostsPostsRootIdIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Posts",
		ParentIdAttr:       "RootId",
		ChildName:          "Posts",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkPostsReactionsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Posts",
		ParentIdAttr: "PostId",
		ChildName:    "Reactions",
		ChildIdAttr:  "",
	})
}

func checkSchemesChannelsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Schemes",
		ParentIdAttr:       "SchemeId",
		ChildName:          "Channels",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkSchemesTeamsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Schemes",
		ParentIdAttr:       "SchemeId",
		ChildName:          "Teams",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkSessionsAuditsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Sessions",
		ParentIdAttr:       "SessionId",
		ChildName:          "Audits",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkTeamsChannelsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	res1 := checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Teams",
		ParentIdAttr: "TeamId",
		ChildName:    "Channels",
		ChildIdAttr:  "Id",
		Filter:       sq.NotEq{"CT.Type": []string{model.CHANNEL_DIRECT, model.CHANNEL_GROUP}},
	})
	res2 := checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Teams",
		ParentIdAttr:       "TeamId",
		ChildName:          "Channels",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
		Filter:             sq.Eq{"CT.Type": []string{model.CHANNEL_DIRECT, model.CHANNEL_GROUP}},
	})
	data1 := res1.Data.(model.RelationalIntegrityCheckData)
	data2 := res2.Data.(model.RelationalIntegrityCheckData)
	data1.Records = append(data1.Records, data2.Records...)
	res1.Data = data1
	return res1
}

func checkTeamsCommandsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Teams",
		ParentIdAttr: "TeamId",
		ChildName:    "Commands",
		ChildIdAttr:  "Id",
	})
}

func checkTeamsIncomingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Teams",
		ParentIdAttr: "TeamId",
		ChildName:    "IncomingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkTeamsOutgoingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Teams",
		ParentIdAttr: "TeamId",
		ChildName:    "OutgoingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkTeamsTeamMembersIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Teams",
		ParentIdAttr: "TeamId",
		ChildName:    "TeamMembers",
		ChildIdAttr:  "",
	})
}

func checkUsersAuditsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Users",
		ParentIdAttr:       "UserId",
		ChildName:          "Audits",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkUsersCommandWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "CommandWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkUsersChannelMemberHistoryIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "ChannelMemberHistory",
		ChildIdAttr:  "",
	})
}

func checkUsersChannelMembersIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "ChannelMembers",
		ChildIdAttr:  "",
	})
}

func checkUsersChannelsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:         "Users",
		ParentIdAttr:       "CreatorId",
		ChildName:          "Channels",
		ChildIdAttr:        "Id",
		CanParentIdBeEmpty: true,
	})
}

func checkUsersCommandsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "CreatorId",
		ChildName:    "Commands",
		ChildIdAttr:  "Id",
	})
}

func checkUsersCompliancesIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Compliances",
		ChildIdAttr:  "Id",
	})
}

func checkUsersEmojiIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "CreatorId",
		ChildName:    "Emoji",
		ChildIdAttr:  "Id",
	})
}

func checkUsersFileInfoIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "CreatorId",
		ChildName:    "FileInfo",
		ChildIdAttr:  "Id",
	})
}

func checkUsersIncomingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "IncomingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkUsersOAuthAccessDataIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "OAuthAccessData",
		ChildIdAttr:  "Token",
	})
}

func checkUsersOAuthAppsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "CreatorId",
		ChildName:    "OAuthApps",
		ChildIdAttr:  "Id",
	})
}

func checkUsersOAuthAuthDataIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "OAuthAuthData",
		ChildIdAttr:  "Code",
	})
}

func checkUsersOutgoingWebhooksIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "CreatorId",
		ChildName:    "OutgoingWebhooks",
		ChildIdAttr:  "Id",
	})
}

func checkUsersPostsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Posts",
		ChildIdAttr:  "Id",
	})
}

func checkUsersPreferencesIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Preferences",
		ChildIdAttr:  "",
	})
}

func checkUsersReactionsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Reactions",
		ChildIdAttr:  "",
	})
}

func checkUsersSessionsIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Sessions",
		ChildIdAttr:  "Id",
	})
}

func checkUsersStatusIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "Status",
		ChildIdAttr:  "",
	})
}

func checkUsersTeamMembersIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "TeamMembers",
		ChildIdAttr:  "",
	})
}

func checkUsersUserAccessTokensIntegrity(ss *SqlStore) model.IntegrityCheckResult {
	return checkParentChildIntegrity(ss, RelationalCheckConfig{
		ParentName:   "Users",
		ParentIdAttr: "UserId",
		ChildName:    "UserAccessTokens",
		ChildIdAttr:  "Id",
	})
}

func checkChannelsIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkChannelsCommandWebhooksIntegrity(ss)
	results <- checkChannelsChannelMemberHistoryIntegrity(ss)
	results <- checkChannelsChannelMembersIntegrity(ss)
	results <- checkChannelsIncomingWebhooksIntegrity(ss)
	results <- checkChannelsOutgoingWebhooksIntegrity(ss)
	results <- checkChannelsPostsIntegrity(ss)
}

func checkCommandsIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkCommandsCommandWebhooksIntegrity(ss)
}

func checkPostsIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkPostsFileInfoIntegrity(ss)
	results <- checkPostsPostsParentIdIntegrity(ss)
	results <- checkPostsPostsRootIdIntegrity(ss)
	results <- checkPostsReactionsIntegrity(ss)
}

func checkSchemesIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkSchemesChannelsIntegrity(ss)
	results <- checkSchemesTeamsIntegrity(ss)
}

func checkSessionsIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkSessionsAuditsIntegrity(ss)
}

func checkTeamsIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkTeamsChannelsIntegrity(ss)
	results <- checkTeamsCommandsIntegrity(ss)
	results <- checkTeamsIncomingWebhooksIntegrity(ss)
	results <- checkTeamsOutgoingWebhooksIntegrity(ss)
	results <- checkTeamsTeamMembersIntegrity(ss)
}

func checkUsersIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	results <- checkUsersAuditsIntegrity(ss)
	results <- checkUsersCommandWebhooksIntegrity(ss)
	results <- checkUsersChannelMemberHistoryIntegrity(ss)
	results <- checkUsersChannelMembersIntegrity(ss)
	results <- checkUsersChannelsIntegrity(ss)
	results <- checkUsersCommandsIntegrity(ss)
	results <- checkUsersCompliancesIntegrity(ss)
	results <- checkUsersEmojiIntegrity(ss)
	results <- checkUsersFileInfoIntegrity(ss)
	results <- checkUsersIncomingWebhooksIntegrity(ss)
	results <- checkUsersOAuthAccessDataIntegrity(ss)
	results <- checkUsersOAuthAppsIntegrity(ss)
	results <- checkUsersOAuthAuthDataIntegrity(ss)
	results <- checkUsersOutgoingWebhooksIntegrity(ss)
	results <- checkUsersPostsIntegrity(ss)
	results <- checkUsersPreferencesIntegrity(ss)
	results <- checkUsersReactionsIntegrity(ss)
	results <- checkUsersSessionsIntegrity(ss)
	results <- checkUsersStatusIntegrity(ss)
	results <- checkUsersTeamMembersIntegrity(ss)
	results <- checkUsersUserAccessTokensIntegrity(ss)
}

func CheckRelationalIntegrity(ss *SqlStore, results chan<- model.IntegrityCheckResult) {
	mlog.Info("Starting relational integrity checks...")
	checkChannelsIntegrity(ss, results)
	checkCommandsIntegrity(ss, results)
	checkPostsIntegrity(ss, results)
	checkSchemesIntegrity(ss, results)
	checkSessionsIntegrity(ss, results)
	checkTeamsIntegrity(ss, results)
	checkUsersIntegrity(ss, results)
	mlog.Info("Done with relational integrity checks")
	close(results)
}
