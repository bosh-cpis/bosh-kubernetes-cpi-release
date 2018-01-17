package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

type Director interface {
	CronItems() ([]CronItem, []error)
	RemoveCronItem(CronItem) error

	UpdateCronItemStatus(CronItemStatus) error
	UpdateCronStatus(CronStatus) error

	Deployments() ([]boshdir.Deployment, error)
	RunErrand(depName, name string, keepAlive, whenChanged bool,
		slugs []boshdir.InstanceGroupOrInstanceSlug) ([]boshdir.ErrandResult, error)
	CleanUp() error
}

var _ Director = DirectorImpl{}
