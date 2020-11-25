package app

import (
	"github.com/masterhung0112/hk_server/model"
)

const (
	TRIGGERWORDS_EXACT_MATCH = 0
	TRIGGERWORDS_STARTS_WITH = 1

	MaxIntegrationResponseSize = 1024 * 1024 // Posts can be <100KB at most, so this is likely more than enough
)

func (a *App) handleWebhookEvents(post *model.Post, team *model.Team, channel *model.Channel, user *model.User) *model.AppError {
	// if !*a.Config().ServiceSettings.EnableOutgoingWebhooks {
	// 	return nil
	// }

	// if channel.Type != model.CHANNEL_OPEN {
	// 	return nil
	// }

	// hooks, err := a.Srv().Store.Webhook().GetOutgoingByTeam(team.Id, -1, -1)
	// if err != nil {
	// 	return model.NewAppError("handleWebhookEvents", "app.webhooks.get_outgoing_by_team.app_error", nil, err.Error(), http.StatusInternalServerError)
	// }

	// if len(hooks) == 0 {
	// 	return nil
	// }

	// var firstWord, triggerWord string

	// splitWords := strings.Fields(post.Message)
	// if len(splitWords) > 0 {
	// 	firstWord = splitWords[0]
	// }

	// relevantHooks := []*model.OutgoingWebhook{}
	// for _, hook := range hooks {
	// 	if hook.ChannelId == post.ChannelId || len(hook.ChannelId) == 0 {
	// 		if hook.ChannelId == post.ChannelId && len(hook.TriggerWords) == 0 {
	// 			relevantHooks = append(relevantHooks, hook)
	// 			triggerWord = ""
	// 		} else if hook.TriggerWhen == TRIGGERWORDS_EXACT_MATCH && hook.TriggerWordExactMatch(firstWord) {
	// 			relevantHooks = append(relevantHooks, hook)
	// 			triggerWord = hook.GetTriggerWord(firstWord, true)
	// 		} else if hook.TriggerWhen == TRIGGERWORDS_STARTS_WITH && hook.TriggerWordStartsWith(firstWord) {
	// 			relevantHooks = append(relevantHooks, hook)
	// 			triggerWord = hook.GetTriggerWord(firstWord, false)
	// 		}
	// 	}
	// }

	// for _, hook := range relevantHooks {
	// 	payload := &model.OutgoingWebhookPayload{
	// 		Token:       hook.Token,
	// 		TeamId:      hook.TeamId,
	// 		TeamDomain:  team.Name,
	// 		ChannelId:   post.ChannelId,
	// 		ChannelName: channel.Name,
	// 		Timestamp:   post.CreateAt,
	// 		UserId:      post.UserId,
	// 		UserName:    user.Username,
	// 		PostId:      post.Id,
	// 		Text:        post.Message,
	// 		TriggerWord: triggerWord,
	// 		FileIds:     strings.Join(post.FileIds, ","),
	// 	}
	// 	a.Srv().Go(func(hook *model.OutgoingWebhook) func() {
	// 		return func() {
	// 			a.TriggerWebhook(payload, hook, post, channel)
	// 		}
	// 	}(hook))
	// }

	return nil
}
