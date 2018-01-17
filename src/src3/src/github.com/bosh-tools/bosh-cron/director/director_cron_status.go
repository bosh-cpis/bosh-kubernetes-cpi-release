package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

func (d DirectorImpl) UpdateCronStatus(status CronStatus) error {
	bytes, err := yaml.Marshal(status)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling cron status")
	}

	return d.director.UpdateConfig("cron-status", "default", bytes)
}
