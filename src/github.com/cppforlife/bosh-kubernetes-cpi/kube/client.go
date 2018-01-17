package kube

import (
	"k8s.io/client-go/kubernetes"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Client struct {
	clientset *kubernetes.Clientset
	namespace string
}

func NewClient(clientset *kubernetes.Clientset, namespace string) Client {
	return Client{clientset, namespace}
}

func (c Client) PVCs() kcorev1.PersistentVolumeClaimInterface {
	return c.clientset.CoreV1().PersistentVolumeClaims(c.namespace)
}

func (c Client) Pods() kcorev1.PodInterface {
	return c.clientset.CoreV1().Pods(c.namespace)
}

func (c Client) ConfigMaps() kcorev1.ConfigMapInterface {
	return c.clientset.CoreV1().ConfigMaps(c.namespace)
}

func (c Client) Services() kcorev1.ServiceInterface {
	return c.clientset.CoreV1().Services(c.namespace)
}
