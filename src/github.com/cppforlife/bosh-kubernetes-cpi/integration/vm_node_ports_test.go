package integration_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("VMs (node_ports cloud property)", func() {
	Describe("create_vm", func() {
		It("can set up vm with node ports", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			nodePortsProperty := []map[string]interface{}{
				map[string]interface{}{
					"name":      "test-port-80",
					"protocol":  "TCP",
					"port":      80,
					"node_port": 32500,
				},
				map[string]interface{}{
					"name":      "test-port-81",
					"protocol":  "UDP",
					"port":      81,
					"node_port": 32501,
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
							"node_ports": nodePortsProperty,
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
					KubeObj{svcs.Items[0].Spec}.Remove("clusterIP").ExpectToEqual(map[string]interface{}{
						"type": "NodePort",

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"externalTrafficPolicy": "Cluster",
						"sessionAffinity":       "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "test-port-80",
								"protocol":   "TCP",
								"port":       80,
								"targetPort": 80,
								"nodePort":   32500,
							},
						},
					})

					KubeObj{svcs.Items[1].Spec}.Remove("clusterIP").ExpectToEqual(map[string]interface{}{
						"type": "NodePort",

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"externalTrafficPolicy": "Cluster",
						"sessionAffinity":       "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "test-port-81",
								"protocol":   "UDP",
								"port":       81,
								"targetPort": 81,
								"nodePort":   32501,
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
