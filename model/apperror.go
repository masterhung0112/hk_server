package model

import (
  "encoding/json"
)

type AppError struct {
  Id                string  `json:"id"`
  Message           string  `json:"message"`
  DetailedError     string  `json:"detailed_error"`
  RequestId         string `json:"request_id,omitempty"`  // The RequestId that's also set in the header
  StatusCode        int     `json:"status_code,omitempty"`
}

func (er *AppError) ToJson() string {
	b, _ := json.Marshal(er)
	return string(b)
}