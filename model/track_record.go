package model

import (
	"net/http"
)

type TrackRecord struct {
	Id                    string
	UserId                string
	Categories            StringArray
	CreateAt              int64
	StartAt               int64
	EndAt                 int64
	WeightedAverage       float64
	WeightedAverageLastId string
	WeightedAverageIsLast bool
}

func (o *TrackRecord) PreSave() {
	if o.Id == "" {
		o.Id = NewId()
	}

	if o.CreateAt == 0 {
		o.CreateAt = GetMillis()
	}

	if o.Categories == nil {
		o.Categories = []string{}
	}

	o.Categories = RemoveDuplicateStrings(o.Categories)
}

func (o *TrackRecord) IsValidWithoutId() error {
	if !(len(o.UserId) == 26 || len(o.UserId) == 0) {
		return NewAppError("TrackRecord.IsValid", "model.trackrecord.is_valid.user_id.app_error", nil, "", http.StatusBadRequest)
	}

	return nil
}
