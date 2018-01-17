package job

import (
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"

	"github.com/bosh-tools/bosh-cron/director"
)

type ErrandJob struct {
	item     director.CronItem
	director director.Director

	logTag string
	logger boshlog.Logger
}

func NewErrandJob(
	item director.CronItem,
	director director.Director,
	logger boshlog.Logger,
) ErrandJob {
	return ErrandJob{
		item:     item,
		director: director,

		logTag: "job.ErrandJob",
		logger: logger,
	}
}

func (s ErrandJob) Run() director.CronItemStatus {
	status := director.CronItemStatus{
		Successful: true,
		Errand:     &director.CronItemStatusErrand{},
	}

	depNames, err := s.depNames()
	if err != nil {
		status.Successful = false
		status.Error = err.Error()
		return status
	}

	s.logger.Debug(s.logTag, "Seleted deployment names '%v'", depNames)

	doneCh := make(chan director.CronItemStatusErrandRun, len(depNames))

	// todo deployment selection relative to team scoped configs?
	for _, depName := range depNames {
		go func(depName string) {
			defer s.logger.HandlePanic("Running errand")
			doneCh <- s.runErrand(depName)
		}(depName)
	}

	for i := 0; i < len(depNames); i++ {
		runStatus := <-doneCh

		status.Errand.Runs = append(status.Errand.Runs, runStatus)

		if !runStatus.Successful {
			status.Successful = false
			status.Error = "Last error: " + runStatus.Error
		}
	}

	return status
}

func (s ErrandJob) runErrand(depName string) director.CronItemStatusErrandRun {
	status := director.CronItemStatusErrandRun{
		Deployment: depName,
		TaskID:     "",
		Successful: true,
	}

	t1 := time.Now()
	results, runErr := s.director.RunErrand(
		depName,
		s.item.Errand.Name,
		s.item.Errand.KeepAlive,
		s.item.Errand.WhenChanged,
		s.item.Errand.InstanceSlugs(),
	)
	t2 := time.Now()

	status.StartedAt = t1.UTC().Format(time.RFC3339)
	status.FinishedAt = t2.UTC().Format(time.RFC3339)

	if runErr != nil {
		status.Successful = false
		status.Error = runErr.Error()
	}

	for _, result := range results {
		if result.ExitCode != 0 {
			status.Successful = false
		}
		status.Results = append(status.Results, director.CronItemStatusErrandResult{
			ExitCode: result.ExitCode,
		})
	}

	return status
}

func (s ErrandJob) depNames() ([]string, error) {
	includeDepNames := s.item.Include.Deployments

	if len(includeDepNames) == 0 {
		deps, err := s.director.Deployments()
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			includeDepNames = append(includeDepNames, dep.Name())
		}
	}

	var depNames []string

	for _, depName := range includeDepNames {
		var excluded bool
		for _, depName2 := range s.item.Exclude.Deployments {
			if depName == depName2 {
				excluded = true
				break
			}
		}
		if !excluded {
			depNames = append(depNames, depName)
		}
	}

	return depNames, nil
}
