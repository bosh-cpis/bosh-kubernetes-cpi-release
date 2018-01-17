package main

import (
	"reflect"
	"time"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"gopkg.in/robfig/cron.v2"

	"github.com/bosh-tools/bosh-cron/director"
	bcjob "github.com/bosh-tools/bosh-cron/job"
)

type Scheduler struct {
	director           director.Director
	checkConfigIterval time.Duration

	currCronItems []director.CronItem

	// https://godoc.org/gopkg.in/robfig/cron.v2
	currCron *cron.Cron
	cronIDs  map[string]cron.EntryID

	logTag string
	logger boshlog.Logger
}

func NewScheduler(director director.Director, logger boshlog.Logger) *Scheduler {
	return &Scheduler{
		director:           director,
		checkConfigIterval: 1 * time.Minute,

		currCron: cron.New(),
		cronIDs:  map[string]cron.EntryID{},

		logTag: "Scheduler",
		logger: logger,
	}
}

func (s *Scheduler) Run() error {
	s.logger.Debug(s.logTag, "Starting scheduler with interval '%s'", s.checkConfigIterval)

	ticker := time.NewTicker(s.checkConfigIterval)

	s.currCron.Start()

	for {
		select {
		case <-ticker.C:
			s.reload()
		}
	}

	return nil
}

func (s *Scheduler) reload() {
	s.logger.Debug(s.logTag, "Reloading cron items")

	defer func() {
		s.logger.Debug(s.logTag, "Finished reloading cron items")
	}()

	var updateErr error

	newCronItems, itemsErrs := s.director.CronItems()
	if len(itemsErrs) == 0 {
		updateErr = s.updateCron(newCronItems)
	}

	status := director.CronStatus{
		ReloadedAt: time.Now().UTC().Format(time.RFC3339),
		Successful: len(itemsErrs) == 0 && updateErr == nil,

		Items: s.fetchStatusItems(),
	}

	for _, err := range itemsErrs {
		status.Errors = append(status.Errors, err.Error())
	}

	if updateErr != nil {
		status.Errors = append(status.Errors, updateErr.Error())
	}

	statusErr := s.director.UpdateCronStatus(status)
	if statusErr != nil {
		s.logger.Error(s.logTag, "Failed updating cron status: %s", statusErr)
	}

	if len(status.Errors) > 0 {
		s.logger.Debug(s.logTag, "Reloaded cron items with errors '%#v'", status.Errors)
	}
}

func (s *Scheduler) fetchStatusItems() []director.CronStatusItem {
	var statuses []director.CronStatusItem

	for _, item := range s.currCronItems {
		entry := s.currCron.Entry(s.cronIDs[item.ID()])

		statuses = append(statuses, director.CronStatusItem{
			Name:       item.Name,
			NextAt:     entry.Next.UTC().Format(time.RFC3339),
			PreviousAt: entry.Prev.UTC().Format(time.RFC3339),
		})
	}

	return statuses
}

func (s *Scheduler) updateCron(newCronItems []director.CronItem) error {
	if reflect.DeepEqual(s.currCronItems, newCronItems) {
		s.logger.Debug(s.logTag, "No changes")
		return nil
	}

	s.logger.Debug(s.logTag, "Detected changes (len %d vs len %d or diff)",
		len(s.currCronItems), len(newCronItems))

	currItemsByID := map[string]director.CronItem{}
	for _, item := range s.currCronItems {
		currItemsByID[item.ID()] = item
	}

	newItemsByID := map[string]director.CronItem{}
	for _, item := range newCronItems {
		newItemsByID[item.ID()] = item
	}

	var errs []error

	for id, item := range currItemsByID {
		if newItem, found := newItemsByID[id]; found {
			if !reflect.DeepEqual(item, newItem) {
				s.logger.Debug(s.logTag, "Unscheduling cron item '%s' to reschedule", item.Name)
				s.currCron.Remove(s.cronIDs[id])
				cronID, err := s.scheduleItem(item)
				if err != nil {
					errs = append(errs, err)
				} else {
					s.cronIDs[id] = cronID
				}
			}
		} else {
			s.logger.Debug(s.logTag, "Unscheduling cron item '%s'", item.Name)
			s.currCron.Remove(s.cronIDs[id])
			delete(s.cronIDs, id)
		}
		delete(newItemsByID, id)
	}

	for id, item := range newItemsByID {
		cronID, err := s.scheduleItem(item)
		if err != nil {
			errs = append(errs, err)
		} else {
			s.cronIDs[id] = cronID
		}
	}

	s.currCronItems = newCronItems

	if len(errs) > 0 {
		return errs[0]
	}

	return nil
}

func (s *Scheduler) scheduleItem(item director.CronItem) (cron.EntryID, error) {
	s.logger.Debug(s.logTag, "Scheduling cron item '%s'", item.Name)

	var job bcjob.Job

	switch {
	case item.Errand != nil:
		job = bcjob.NewErrandJob(item, s.director, s.logger)
	case item.Cleanup != nil:
		job = bcjob.NewCleanupJob(item, s.director, s.logger)
	default:
		return cron.EntryID(-1), bosherr.Errorf("Unknown cron item '%#v'", item)
	}

	if item.RunOnce {
		job = bcjob.NewRunOnceJob(item, job, s.director, s.logger)
	}

	jobWrapper := bcjob.NewStatusWrapper(item, job, s.director, s.logger)

	cronID, err := s.currCron.AddJob(item.Schedule, jobWrapper)
	if err != nil {
		return cron.EntryID(-1), bosherr.WrapErrorf(err, "Adding cron item '%s' to cron", item.Name)
	}

	return cronID, nil
}
