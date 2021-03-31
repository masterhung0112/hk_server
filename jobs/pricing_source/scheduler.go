package pricing_source

import (
	"time"

	"github.com/masterhung0112/hk_server/v5/app"
	"github.com/masterhung0112/hk_server/v5/model"
)


const (
	SchedFreqSeconds = 60
)

type Scheduler struct {
	App *app.App
}

func (m *PricingSourceJobInterfaceImpl) MakeScheduler() model.Scheduler {
	return &Scheduler{m.App}
}

func (scheduler *Scheduler) Name() string {
	return JobName + "Scheduler"
}

func (scheduler *Scheduler) JobType() string {
	return model.JOB_TYPE_PRICING_SOURCE
}

func (scheduler *Scheduler) Enabled(cfg *model.Config) bool {
	// Only enabled when Metrics are enabled.
	return *cfg.MetricsSettings.Enable
}

func (scheduler *Scheduler) NextScheduleTime(cfg *model.Config, now time.Time, pendingJobs bool, lastSuccessfulJob *model.Job) *time.Time {
	nextTime := time.Now().Add(SchedFreqSeconds * time.Second)
	return &nextTime
}

func (scheduler *Scheduler) ScheduleJob(cfg *model.Config, pendingJobs bool, lastSuccessfulJob *model.Job) (*model.Job, *model.AppError) {
	data := map[string]string{}

	job, err := scheduler.App.Srv().Jobs.CreateJob(model.JOB_TYPE_PRICING_SOURCE, data)
	if err != nil {
		return nil, err
	}
	return job, nil
}