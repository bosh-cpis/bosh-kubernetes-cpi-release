package job

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/bosh-tools/bosh-cron/director"
)

type StatusWrapper struct {
	item     director.CronItem
	job      Job
	director director.Director

	logTag string
	logger boshlog.Logger
}

func NewStatusWrapper(
	item director.CronItem,
	job Job,
	director director.Director,
	logger boshlog.Logger,
) StatusWrapper {
	return StatusWrapper{
		item:     item,
		job:      job,
		director: director,

		logTag: "job.StatusWrapper",
		logger: logger,
	}
}

func (s StatusWrapper) Run() {
	s.logger.Debug(s.logTag, "Started running cron item '%s'", s.item.Name)

	defer func() {
		s.logger.Debug(s.logTag, "Finished running cron item '%s'", s.item.Name)
	}()

	t1 := time.Now()
	status := s.job.Run()
	t2 := time.Now()

	status.Name = s.item.Name
	status.StartedAt = t1.UTC().Format(time.RFC3339)
	status.FinishedAt = t2.UTC().Format(time.RFC3339)

	reportErr := s.director.UpdateCronItemStatus(status)
	if reportErr != nil {
		logFmt := "Failed to report cron item '%s' status '%#v': %s"
		s.logger.Error(s.logTag, logFmt, s.item.Name, status, reportErr)
	}
}
