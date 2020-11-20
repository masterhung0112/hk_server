package model

type SlackAttachment struct {
	Id         int64                   `json:"id"`
	Fallback   string                  `json:"fallback"`
	Color      string                  `json:"color"`
	Pretext    string                  `json:"pretext"`
	AuthorName string                  `json:"author_name"`
	AuthorLink string                  `json:"author_link"`
	AuthorIcon string                  `json:"author_icon"`
	Title      string                  `json:"title"`
	TitleLink  string                  `json:"title_link"`
	Text       string                  `json:"text"`
	Fields     []*SlackAttachmentField `json:"fields"`
	ImageURL   string                  `json:"image_url"`
	ThumbURL   string                  `json:"thumb_url"`
	Footer     string                  `json:"footer"`
	FooterIcon string                  `json:"footer_icon"`
	Timestamp  interface{}             `json:"ts"` // This is either a string or an int64
	Actions    []*PostAction           `json:"actions,omitempty"`
}

func (s *SlackAttachment) Equals(input *SlackAttachment) bool {
	// Direct comparison of simple types

	if s.Id != input.Id {
		return false
	}

	if s.Fallback != input.Fallback {
		return false
	}

	if s.Color != input.Color {
		return false
	}

	if s.Pretext != input.Pretext {
		return false
	}

	if s.AuthorName != input.AuthorName {
		return false
	}

	if s.AuthorLink != input.AuthorLink {
		return false
	}

	if s.AuthorIcon != input.AuthorIcon {
		return false
	}

	if s.Title != input.Title {
		return false
	}

	if s.TitleLink != input.TitleLink {
		return false
	}

	if s.Text != input.Text {
		return false
	}

	if s.ImageURL != input.ImageURL {
		return false
	}

	if s.ThumbURL != input.ThumbURL {
		return false
	}

	if s.Footer != input.Footer {
		return false
	}

	if s.FooterIcon != input.FooterIcon {
		return false
	}

	// Compare length & slice values of fields
	if len(s.Fields) != len(input.Fields) {
		return false
	}

	for j := range s.Fields {
		if !s.Fields[j].Equals(input.Fields[j]) {
			return false
		}
	}

	// Compare length & slice values of actions
	if len(s.Actions) != len(input.Actions) {
		return false
	}

	for j := range s.Actions {
		if !s.Actions[j].Equals(input.Actions[j]) {
			return false
		}
	}

	return s.Timestamp == input.Timestamp
}

type SlackAttachmentField struct {
	Title string              `json:"title"`
	Value interface{}         `json:"value"`
	Short SlackCompatibleBool `json:"short"`
}

func (s *SlackAttachmentField) Equals(input *SlackAttachmentField) bool {
	if s.Title != input.Title {
		return false
	}

	if s.Value != input.Value {
		return false
	}

	if s.Short != input.Short {
		return false
	}

	return true
}
