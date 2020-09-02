package model

// Session contains the user session details.
// This struct's serializer methods are auto-generated. If a new field is added/removed,
// please run make gen-serialized.
type Session struct {
  Id             string        `json:"id"`
  UserId         string        `json:"user_id"`
}