package job

import (
	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/bosh-tools/bosh-cron/director"
)

type CleanupJob struct {
	item     director.CronItem
	director director.Director

	logTag string
	logger boshlog.Logger
}

func NewCleanupJob(
	item director.CronItem,
	director director.Director,
	logger boshlog.Logger,
) CleanupJob {
	return CleanupJob{
		item:     item,
		director: director,

		logTag: "job.CleanupJob",
		logger: logger,
	}
}

func (s CleanupJob) Run() director.CronItemStatus {
	runErr := s.director.CleanUp()

	status := director.CronItemStatus{
		Successful: true,
		Cleanup: &director.CronItemStatusCleanup{
			TaskID: "",
		},
	}

	if runErr != nil {
		status.Successful = false
		status.Error = runErr.Error()
	}

	return status
}
