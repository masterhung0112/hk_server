package model

type User struct {
  Id          string      `json:"id"`
  CreatedAt   int64       `json:"created_at,omitempty"`
  UpdatedAt   int64       `json:"updated_at,omitempty"`
  DeletedAt   int64       `json:"deleted_at"`
  UserName    string      `json:"username"`
  Password    string      `json:"password,omitempty"`
  Email       string      `json:"email"`
  Roles       string      `json:"roles"`
}