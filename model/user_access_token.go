package model

type UserAccessToken struct {
	Id          string `json:"id"`
	Token       string `json:"token,omitempty"`
	UserId      string `json:"user_id"`
	Description string `json:"description"`
	IsActive    bool   `json:"is_active"`
}

func (t *UserAccessToken) PreSave() {
	t.Id = NewId()
	t.IsActive = true
}
