package model

import (
  "encoding/json"
  "io"
)

type User struct {
  Id                    string      `json:"id"`
  CreateAt              int64       `json:"create_at,omitempty"`
  UpdateAt              int64       `json:"update_at,omitempty"`
  DeleteAt              int64       `json:"delete_at"`
  Username              string      `json:"username"`
  Password              string      `json:"password,omitempty"`
  Email                 string      `json:"email"`
  EmailVerified         bool      `json:"email_verified,omitempty"`
  FirstName             string      `json:"first_name"`
	LastName              string      `json:"last_name"`
  Roles                 string      `json:"roles"`
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