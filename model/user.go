package model

import (
  "encoding/json"
  "io"
)

type User struct {
  Id          string      `json:"id"`
  CreatedAt   int64       `json:"created_at,omitempty"`
  UpdatedAt   int64       `json:"updated_at,omitempty"`
  DeletedAt   int64       `json:"deleted_at"`
  Username    string      `json:"username"`
  Password    string      `json:"password,omitempty"`
  Email       string      `json:"email"`
  FirstName   string      `json:"first_name"`
	LastName    string      `json:"last_name"`
  Roles       string      `json:"roles"`
}

// UserFromJson will decode the input and return a User
func UserFromJson(data io.Reader) *User {
  var user *User
  json.NewDecoder(data).Decode(&user)
  return user
}

func (user *User) ToJson() string {
  b, _ := json.Marshal(user)
  return string(b)
}