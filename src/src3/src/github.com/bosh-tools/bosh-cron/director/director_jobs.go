package director

import (
	boshdir "github.com/cloudfoundry/bosh-cli/director"
)

func (d DirectorImpl) Deployments() ([]boshdir.Deployment, error) {
	return d.director.Deployments()
}

func (d DirectorImpl) RunErrand(
	depName, name string, keepAlive, whenChanged bool,
	slugs []boshdir.InstanceGroupOrInstanceSlug) ([]boshdir.ErrandResult, error) {

	dep, err := d.director.FindDeployment(depName)
	if err != nil {
		return nil, err
	}

	return dep.RunErrand(name, keepAlive, whenChanged, slugs)
}

func (d DirectorImpl) CleanUp() error {
	return d.director.CleanUp(false)
}
