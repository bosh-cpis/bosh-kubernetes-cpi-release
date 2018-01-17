package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("VMs (cluster_ips cloud property)", func() {
	Describe("create_vm", func() {
		It("can set up vm with cluster IPs", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			clusterIPsProperty := []map[string]interface{}{
				map[string]interface{}{
					"name":       "test-ip-0",
					"cluster_ip": "10.96.0.100",
					"ports": []map[string]interface{}{
						map[string]interface{}{
							"name":     "test-port-80",
							"protocol": "TCP",
							"port":     80,
						},
						map[string]interface{}{
							"name":     "test-port-81",
							"protocol": "UDP",
							"port":     81,
						},
					},
				},
				map[string]interface{}{
					"name":       "test-ip-1",
					"cluster_ip": "10.96.0.101",
					"ports": []map[string]interface{}{
						map[string]interface{}{
							"name":     "test-port-90",
							"protocol": "TCP",
							"port":     90,
						},
						map[string]interface{}{
							"name":     "test-port-91",
							"protocol": "UDP",
							"port":     91,
						},
					},
				},
			}

			// Try twice to check for collisions on cleaned up resources
			for range []int{0, 1} {
				resp := cpi.Exec(testlib.CPIRequest{
					Method: "create_vm",
					Arguments: []interface{}{
						"agent-id",
						"scl-123/" + props.DummyImage(),
						map[string]interface{}{
							"cluster_ips": clusterIPsProperty,
						},
						props.RandomNetwork(),
						nil,
						map[string]interface{}{"env-key": "env-val"},
					},
				})

				listOpts := kmetav1.ListOptions{LabelSelector: "bosh.io/vm-cid=" + resp.StringResult()}

				By("checking created pod", func() {
					pods, err := cpi.Kube().Pods().List(listOpts)
					Expect(err).ToNot(HaveOccurred())

					Expect(pods.Items).To(HaveLen(1))
					Expect(pods.Items[0].ObjectMeta.Name).To(Equal(resp.StringResult()))
				})

				By("checking created servcies for selected for the pod", func() {
					svcs, err := cpi.Kube().Services().List(listOpts)
					Expect(err).ToNot(HaveOccurred())

					Expect(svcs.Items).To(HaveLen(2))
					KubeObj{svcs.Items[0].Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": "10.96.0.100",

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "test-port-80",
								"protocol":   "TCP",
								"port":       80,
								"targetPort": 80,
							},
							map[string]interface{}{
								"name":       "test-port-81",
								"protocol":   "UDP",
								"port":       81,
								"targetPort": 81,
							},
						},
					})

					KubeObj{svcs.Items[1].Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": "10.96.0.101",

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "test-port-90",
								"protocol":   "TCP",
								"port":       90,
								"targetPort": 90,
							},
							map[string]interface{}{
								"name":       "test-port-91",
								"protocol":   "UDP",
								"port":       91,
								"targetPort": 91,
							},
						},
					})
				})

				By("deleting it", func() {
					cpi.Exec(testlib.CPIRequest{
						Method:    "delete_vm",
						Arguments: []interface{}{resp.StringResult()},
					})
				})

				By("checking pod is deleted", func() {
					_, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				By("checking services are deleted", func() {
					svcs, err := cpi.Kube().Services().List(listOpts)
					Expect(err).ToNot(HaveOccurred())
					Expect(svcs.Items).To(BeEmpty())
				})
			}
		})
	})
})
