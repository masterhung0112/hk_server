package model

type TrackRecord struct {
  Id                    string
  UserId                string
	Category              []string
	CreateAt              int64
	StartAt               int64
	EndAt                 int64
	Duration              int64
	WeightedAverage       float64
	WeightedAverageLastId string
	WeightedAverageIsLast bool
}
