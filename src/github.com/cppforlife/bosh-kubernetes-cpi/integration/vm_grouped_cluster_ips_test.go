package integration_test

import (
	"fmt"
	"runtime/debug"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("VMs (grouped cluster_ips cloud property)", func() {
	Describe("create_vm", func() {
		It("can set up vm with cluster IPs for multiple pods", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			clusterIPsProperty := []map[string]interface{}{
				map[string]interface{}{
					"name":       "test-ip-0",
					"cluster_ip": "10.96.0.100",
					"grouped":    true,
					"ports": []map[string]interface{}{
						map[string]interface{}{
							"name":     "test-port-80",
							"protocol": "TCP",
							"port":     80,
						},
					},
				},
			}

			vmEnv := map[string]interface{}{
				"bosh": map[string]interface{}{
					"group": "test-env-bosh-group1",
				},
			}

			sendPanic := func(ch chan interface{}) {
				obj := recover()
				if obj != nil {
					fmt.Printf("-----> Paniced: %s\n", debug.Stack())
					ch <- obj
				}
			}

			// Try twice to check for collisions on cleaned up resources
			for range []int{0, 1} {
				var resp, resp2 testlib.CPIResponse

				respDone := make(chan interface{})
				resp2Done := make(chan interface{})

				go func() {
					defer sendPanic(respDone)
					resp = cpi.Exec(testlib.CPIRequest{
						Method: "create_vm",
						Arguments: []interface{}{
							"agent-id",
							"scl-123/" + props.DummyImage(),
							map[string]interface{}{
								"cluster_ips": clusterIPsProperty,
							},
							props.RandomNetwork(),
							nil,
							vmEnv,
						},
					})
					respDone <- nil
				}()

				go func() {
					defer sendPanic(resp2Done)
					resp2 = cpi.Exec(testlib.CPIRequest{
						Method: "create_vm",
						Arguments: []interface{}{
							"agent-id",
							"scl-123/" + props.DummyImage(),
							map[string]interface{}{
								"cluster_ips": clusterIPsProperty,
							},
							props.RandomNetwork(),
							nil,
							vmEnv,
						},
					})
					resp2Done <- nil
				}()

				Expect(<-respDone).To(BeNil())
				Expect(<-resp2Done).To(BeNil())

				listOpts := kmetav1.ListOptions{LabelSelector: "bosh.io/group=test-env-bosh-group1"}

				By("checking created pods", func() {
					pods, err := cpi.Kube().Pods().List(listOpts)
					Expect(err).ToNot(HaveOccurred())

					Expect(pods.Items).To(
						HaveLen(2), "Expected to find only 2 pods (may be left overs from previous runs?)")
					Expect([]string{pods.Items[0].ObjectMeta.Name, pods.Items[1].ObjectMeta.Name}).To(
						ConsistOf(resp.StringResult(), resp2.StringResult()))
				})

				By("checking created servcies for selected for the pod", func() {
					svcs, err := cpi.Kube().Services().List(listOpts)
					Expect(err).ToNot(HaveOccurred())

					Expect(svcs.Items).To(HaveLen(1))
					KubeObj{svcs.Items[0].Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": "10.96.0.100",

						"selector": map[string]interface{}{
							"bosh.io/group": "test-env-bosh-group1",
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "test-port-80",
								"protocol":   "TCP",
								"port":       80,
								"targetPort": 80,
							},
						},
					})
				})

				By("deleting pods", func() {
					go func() {
						defer sendPanic(respDone)
						cpi.Exec(testlib.CPIRequest{
							Method:    "delete_vm",
							Arguments: []interface{}{resp.StringResult()},
						})
						respDone <- nil
					}()

					go func() {
						defer sendPanic(resp2Done)
						cpi.Exec(testlib.CPIRequest{
							Method:    "delete_vm",
							Arguments: []interface{}{resp2.StringResult()},
						})
						resp2Done <- nil
					}()

					Expect(<-respDone).To(BeNil())
					Expect(<-resp2Done).To(BeNil())
				})

				By("checking pod is deleted", func() {
					_, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())

					_, err = cpi.Kube().Pods().Get(resp2.StringResult(), kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				By("checking services are not deleted", func() {
					svcs, err := cpi.Kube().Services().List(listOpts)
					Expect(err).ToNot(HaveOccurred())
					Expect(svcs.Items).To(HaveLen(1))
				})
			}

			By("cleaning up cluster IP service", func() {
				err := cpi.Kube().Services().Delete("test-env-bosh-group1-test-ip-0", kmetav1.NewDeleteOptions(0))
				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
