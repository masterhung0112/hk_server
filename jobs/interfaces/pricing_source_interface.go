package interfaces

import (
	"github.com/masterhung0112/hk_server/v5/model"
)

type PricingSourceJobInterface interface {
	MakeWorker() model.Worker
	MakeScheduler() model.Scheduler
}
