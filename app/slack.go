package app

import (
  "fmt"
	"github.com/masterhung0112/hk_server/model"
)

func (a *App) ProcessSlackText(text string) string {
	//TODO: Open
	// text = expandAnnouncement(text)
	// text = replaceUserIds(a.Srv().Store.User(), text)

	return text
}

// Expand announcements in incoming webhooks from Slack. Those announcements
// can be found in the text attribute, or in the pretext, text, title and value
// attributes of the attachment structure. The Slack attachment structure is
// documented here: https://api.slack.com/docs/attachments
func (a *App) ProcessSlackAttachments(attachments []*model.SlackAttachment) []*model.SlackAttachment {
	var nonNilAttachments = model.StringifySlackFieldValue(attachments)
	for _, attachment := range attachments {
		attachment.Pretext = a.ProcessSlackText(attachment.Pretext)
		attachment.Text = a.ProcessSlackText(attachment.Text)
		attachment.Title = a.ProcessSlackText(attachment.Title)

		for _, field := range attachment.Fields {
			if field.Value != nil {
				// Ensure the value is set to a string if it is set
				field.Value = a.ProcessSlackText(fmt.Sprintf("%v", field.Value))
			}
		}
	}
	return nonNilAttachments
}
