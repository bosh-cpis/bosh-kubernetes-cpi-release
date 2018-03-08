package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	boshdir "github.com/cloudfoundry/bosh-cli/director"
	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	kapiv1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Instance struct {
	instance   boshdir.Instance
	deployment boshdir.Deployment
	director   boshdir.Director
	podsClient kcorev1.PodInterface
}

func NewInstance(
	instance boshdir.Instance,
	deployment boshdir.Deployment,
	director boshdir.Director,
	podsClient kcorev1.PodInterface,
) Instance {
	return Instance{instance, deployment, director, podsClient}
}

func (i Instance) ResurrectIfNecessary() error {
	if len(i.instance.VMID) == 0 {
		return i.Resurrect()
	}

	pod, err := i.podsClient.Get(i.instance.VMID, kmetav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return i.Resurrect()
		} else {
			return bosherr.WrapErrorf(err, "Getting pod")
		}
	}

	if pod.Status.Phase != kapiv1.PodRunning {
		return i.Resurrect()
	}

	return nil
}

func (i Instance) Resurrect() error {
	if len(i.instance.ID) == 0 || len(i.instance.Group) == 0 {
		return bosherr.Errorf("Expected instance to have name and id for ressurrection")
	}

	if !i.instance.ExpectsVM {
		return nil
	}

	path := fmt.Sprintf("/deployments/%s/scan_and_fix", i.deployment.Name())

	body := map[string]interface{}{
		"jobs": map[string]interface{}{
			i.instance.Group: []string{i.instance.ID},
		},
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return bosherr.WrapErrorf(err, "Serializing resurrection request body")
	}

	reqFunc := func(req *http.Request) {
		req.Header.Set("Content-Type", "application/json")
	}

	_, _, err = i.director.(boshdir.DirectorImpl).NewHTTPClientRequest().RawPut(path, bodyBytes, reqFunc)
	if err != nil {
		return bosherr.WrapErrorf(err, "Making resurrection request")
	}

	return nil
}
