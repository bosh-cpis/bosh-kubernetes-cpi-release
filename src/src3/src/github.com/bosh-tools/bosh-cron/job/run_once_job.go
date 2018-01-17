package job

import (
	"sync"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/bosh-tools/bosh-cron/director"
)

type RunOnceJob struct {
	item     director.CronItem
	job      Job
	director director.Director

	once sync.Once

	logTag string
	logger boshlog.Logger
}

func NewRunOnceJob(
	item director.CronItem,
	job Job,
	director director.Director,
	logger boshlog.Logger,
) *RunOnceJob {
	return &RunOnceJob{
		item:     item,
		job:      job,
		director: director,

		logTag: "job.RunOnceJob",
		logger: logger,
	}
}

func (s *RunOnceJob) Run() director.CronItemStatus {
	var status director.CronItemStatus

	s.once.Do(func() {
		status = s.job.Run()

		removeErr := s.director.RemoveCronItem(s.item)
		if removeErr != nil {
			logFmt := "Failed to remove cron item '%s': %s"
			s.logger.Error(s.logTag, logFmt, s.item.Name, removeErr)
		}
	})

	return status
}
