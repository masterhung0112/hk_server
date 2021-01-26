package model

import (
	"encoding/json"
	"io"
	"net/http"
)

type TrackPoint struct {
	Id         string   `json:"id"`
	TargetId   string   `json:"target_id"`
	Point      GeoPoint `json:"point"`
	CreateAt   int64    `json:"create_at"`
  DeviceId   string   `json:"device_id"`
  Duration   int64    `json:"duration"`
}

func (r *TrackPoint) ToJson() string {
	b, _ := json.Marshal(r)
	return string(b)
}

func TrackPointFromJson(data io.Reader) *TrackPoint {
	var r *TrackPoint
	json.NewDecoder(data).Decode(&r)
	return r
}

func (o *TrackPoint) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	o.CreateAt = GetMillis()
}

func (o *TrackPoint) IsValid() *AppError {
	if !IsValidId(o.Id) {
		return NewAppError("TrackPoint.IsValid", "model.trackpoint.is_valid.id.app_error", nil, "", http.StatusBadRequest)
	}

	return o.IsValidWithoutId()
}

func (o *TrackPoint) IsValidWithoutId() *AppError {
	if o.CreateAt == 0 {
		return NewAppError("TrackPoint.IsValid", "model.trackpoint.is_valid.create_at.app_error", nil, "id="+o.Id, http.StatusBadRequest)
	}

	return nil
}
