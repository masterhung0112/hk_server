package model

import (
  "encoding/json"
)

type AppError struct {
  Id                string  `json:"id"`
  Message           string  `json:"message"`
  Where             string  `json:"where"`
  DetailedError     string  `json:"detailed_error"`
  RequestId         string  `json:"request_id,omitempty"`  // The RequestId that's also set in the header
  StatusCode        int     `json:"status_code,omitempty"`
}

func (err *AppError) Error() string {
  return err.Where + ":" + err.Message
}

func (er *AppError) ToJson() string {
  b, _ := json.Marshal(er)
  return string(b)
}