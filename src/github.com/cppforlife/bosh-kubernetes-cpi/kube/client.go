package kube

import (
	"k8s.io/client-go/kubernetes"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	kv1beta1 "k8s.io/client-go/kubernetes/typed/policy/v1beta1"
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

func (c Client) PDBs() kv1beta1.PodDisruptionBudgetInterface {
	return c.clientset.PolicyV1beta1().PodDisruptionBudgets(c.namespace)
}
