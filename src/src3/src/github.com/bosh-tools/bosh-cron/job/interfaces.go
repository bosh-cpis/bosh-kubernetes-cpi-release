package job

import (
	"github.com/bosh-tools/bosh-cron/director"
)

type Job interface {
	Run() director.CronItemStatus
}

var _ Job = ErrandJob{}
var _ Job = CleanupJob{}
var _ Job = &RunOnceJob{}
