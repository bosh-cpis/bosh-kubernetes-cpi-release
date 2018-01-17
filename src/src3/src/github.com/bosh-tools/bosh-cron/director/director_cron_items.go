package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	"gopkg.in/yaml.v2"
)

type DirectorImpl struct {
	director boshdir.Director
}

func (d DirectorImpl) CronItems() ([]CronItem, []error) {
	result, errs := d.individualCronItems()

	// todo namespacing?
	cronItemNames := map[string]struct{}{}

	for _, cronItem := range result {
		if _, found := cronItemNames[cronItem.Name]; found {
			errs = append(errs, bosherr.Errorf(
				"Expected to not find duplicate cron item '%s'", cronItem.Name))
			continue
		}
		cronItemNames[cronItem.Name] = struct{}{}
	}

	return result, errs
}

func (d DirectorImpl) RemoveCronItem(item CronItem) error {
	_, err := d.director.DeleteConfig("cron-item", item.Name)
	return err
}

func (d DirectorImpl) individualCronItems() ([]CronItem, []error) {
	var result []CronItem
	var errs []error

	configItems, err := d.director.ListConfigs(boshdir.ConfigsFilter{Type: "cron-item"})
	if err != nil {
		return nil, []error{err}
	}

	for _, configItem := range configItems {
		config, err := d.director.LatestConfig(configItem.Type, configItem.Name)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err,
				"Fetching config (type: '%s' name: '%s')", configItem.Type, configItem.Name))
			continue
		}

		var cronItem CronItem

		err = yaml.Unmarshal([]byte(config.Content), &cronItem)
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err,
				"Unmarshaling config (type: '%s' name: '%s')", configItem.Type, configItem.Name))
			continue
		}

		cronItem.Name = configItem.Name

		err = cronItem.Validate()
		if err != nil {
			errs = append(errs, bosherr.WrapErrorf(err,
				"Validating config (type: '%s' name: '%s')", configItem.Type, configItem.Name))
			continue
		}

		result = append(result, cronItem)
	}

	return result, errs
}
