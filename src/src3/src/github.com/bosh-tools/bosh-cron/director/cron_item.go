package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
)

type CronItem struct {
	Name string

	Schedule string
	RunOnce  bool `yaml:"run_once"`

	Errand  *CronItemErrand
	Cleanup *CronItemCleanup

	Include CronItemInclude
	Exclude CronItemExclude
}

type CronItemErrand struct {
	Name        string
	Instances   []string
	WhenChanged bool `yaml:"when_changed"`
	KeepAlive   bool `yaml:"keep_alive"`
}

type CronItemCleanup struct{}

type CronItemInclude struct {
	Deployments []string
}

type CronItemExclude struct {
	Deployments []string
}

func (i CronItem) ID() string { return i.Name }

func (i CronItem) Validate() error {
	if len(i.Name) == 0 {
		return bosherr.Error("Missing 'name'")
	}
	if len(i.Schedule) == 0 {
		return bosherr.Error("Missing 'schedule'") // todo validate format
	}

	switch {
	case i.Errand != nil:
		if len(i.Errand.Name) == 0 {
			return bosherr.Error("Validating 'errand': Missing 'name'")
		}
		for idx, instName := range i.Errand.Instances {
			if len(instName) == 0 {
				return bosherr.Errorf("Validating 'errand.instances[%d]': Expected to be non-empty", idx)
			}
		}
		if _, err := i.Errand.instanceSlugsOrErr(); err != nil {
			return bosherr.WrapErrorf(err, "Validating 'errand.instances'")
		}

	case i.Cleanup != nil:
		// nothing to validate

	default:
		return bosherr.Error("Missing 'errand' or 'cleanup'")
	}

	for idx, depName := range i.Include.Deployments {
		if len(depName) == 0 {
			return bosherr.Errorf("Validating 'include.deployments[%d]': Expected to be non-empty", idx)
		}
	}
	return nil
}

func (e CronItemErrand) InstanceSlugs() []boshdir.InstanceGroupOrInstanceSlug {
	slugs, _ := e.instanceSlugsOrErr()
	return slugs
}

func (e CronItemErrand) instanceSlugsOrErr() ([]boshdir.InstanceGroupOrInstanceSlug, error) {
	var result []boshdir.InstanceGroupOrInstanceSlug

	for _, instName := range e.Instances {
		slug, err := boshdir.NewInstanceGroupOrInstanceSlugFromString(instName)
		if err != nil {
			return nil, err
		}
		result = append(result, slug)
	}

	return result, nil
}
