package director

import (
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

func (d DirectorImpl) UpdateCronItemStatus(status CronItemStatus) error {
	bytes, err := yaml.Marshal(status)
	if err != nil {
		return bosherr.WrapErrorf(err, "Marshaling cron item status '%s'", status.Name)
	}

	return d.director.UpdateConfig("cron-item-status", status.Name, bytes)
}
