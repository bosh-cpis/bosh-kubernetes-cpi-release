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

var _ = Describe("VMs (networking)", func() {
	Describe("create_vm", func() {
		It("can set up vm with manual and vip networking", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			ipSvcNames := []string{
				"bosh-ip-" + props.IPStr(props.ManualNetwork1IP()),
				"bosh-ip-" + props.IPStr(props.ManualNetwork2IP()),
				"bosh-ip-" + props.IPStr(props.VIPNetwork1IP()),
				"bosh-ip-" + props.IPStr(props.VIPNetwork2IP()),
			}

			sendPanic := func(ch chan interface{}) {
				obj := recover()
				if obj != nil {
					fmt.Printf("-----> Paniced: %s\n", debug.Stack())
					ch <- obj
				}
			}

			By("checking IPs do not exist", func() {
				for _, svcName := range ipSvcNames {
					_, err := cpi.Kube().Services().Get(svcName, kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				}
			})

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
							"scl-123/" + props.StemcellImage(),
							map[string]interface{}{
								"ingresses": []map[string]interface{}{
									map[string]interface{}{
										"ports": []interface{}{1232, 1233},
									},
									map[string]interface{}{
										"ports":    []interface{}{1234},
										"networks": []interface{}{"private1"},
									},
									map[string]interface{}{
										"ports":    []interface{}{1235},
										"networks": []interface{}{"private2"},
									},
									map[string]interface{}{
										"ports":     []interface{}{1236},
										"protocols": []interface{}{"udp"},
									},
								},
							},
							map[string]interface{}{
								"private1": props.ManualNetwork1(),
								"private2": props.VIPNetwork1(),
							},
							nil,
							nil,
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
							"scl-123/" + props.StemcellImage(),
							map[string]interface{}{
								"ingresses": []map[string]interface{}{
									map[string]interface{}{
										"ports": []interface{}{1232, 1233},
									},
								},
							},
							map[string]interface{}{
								"private1": props.ManualNetwork2(),
								"private2": props.VIPNetwork2(),
							},
							nil,
							nil,
						},
					})
					resp2Done <- nil
				}()

				Expect(<-respDone).To(BeNil())
				Expect(<-resp2Done).To(BeNil())

				By("checking created services for first pod", func() {
					svc, err := cpi.Kube().Services().Get(
						"bosh-ip-"+props.IPStr(props.ManualNetwork1IP()), kmetav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					KubeObj{svc.Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": props.ManualNetwork1IP(),

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "bosh-port-0-0-0",
								"protocol":   "TCP",
								"port":       1232,
								"targetPort": 1232,
							},
							map[string]interface{}{
								"name":       "bosh-port-0-1-0",
								"protocol":   "TCP",
								"port":       1233,
								"targetPort": 1233,
							},
							map[string]interface{}{
								"name":       "bosh-port-1-0-0",
								"protocol":   "TCP",
								"port":       1234,
								"targetPort": 1234,
							},
							map[string]interface{}{
								"name":       "bosh-port-2-0-0",
								"protocol":   "UDP",
								"port":       1236,
								"targetPort": 1236,
							},
						},
					})

					svc, err = cpi.Kube().Services().Get(
						"bosh-ip-"+props.IPStr(props.VIPNetwork1IP()), kmetav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					KubeObj{svc.Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": props.VIPNetwork1IP(),

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "bosh-port-0-0-0",
								"protocol":   "TCP",
								"port":       1232,
								"targetPort": 1232,
							},
							map[string]interface{}{
								"name":       "bosh-port-0-1-0",
								"protocol":   "TCP",
								"port":       1233,
								"targetPort": 1233,
							},
							map[string]interface{}{
								"name":       "bosh-port-1-0-0",
								"protocol":   "TCP",
								"port":       1235,
								"targetPort": 1235,
							},
							map[string]interface{}{
								"name":       "bosh-port-2-0-0",
								"protocol":   "UDP",
								"port":       1236,
								"targetPort": 1236,
							},
						},
					})
				})

				By("checking created services for second pod", func() {
					svc, err := cpi.Kube().Services().Get(
						"bosh-ip-"+props.IPStr(props.ManualNetwork2IP()), kmetav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					KubeObj{svc.Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": props.ManualNetwork2IP(),

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp2.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "bosh-port-0-0-0",
								"protocol":   "TCP",
								"port":       1232,
								"targetPort": 1232,
							},
							map[string]interface{}{
								"name":       "bosh-port-0-1-0",
								"protocol":   "TCP",
								"port":       1233,
								"targetPort": 1233,
							},
						},
					})

					svc, err = cpi.Kube().Services().Get(
						"bosh-ip-"+props.IPStr(props.VIPNetwork2IP()), kmetav1.GetOptions{})
					Expect(err).ToNot(HaveOccurred())

					KubeObj{svc.Spec}.ExpectToEqual(map[string]interface{}{
						"type":      "ClusterIP",
						"clusterIP": props.VIPNetwork2IP(),

						"selector": map[string]interface{}{
							"bosh.io/vm-cid": resp2.StringResult(),
						},

						"sessionAffinity": "None",

						"ports": []map[string]interface{}{
							map[string]interface{}{
								"name":       "bosh-port-0-0-0",
								"protocol":   "TCP",
								"port":       1232,
								"targetPort": 1232,
							},
							map[string]interface{}{
								"name":       "bosh-port-0-1-0",
								"protocol":   "TCP",
								"port":       1233,
								"targetPort": 1233,
							},
						},
					})
				})

				By("running networking test", func() {
					// todo assert on packets
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

				By("checking pods are deleted", func() {
					_, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())

					_, err = cpi.Kube().Pods().Get(resp2.StringResult(), kmetav1.GetOptions{})
					Expect(err).To(HaveOccurred())
					Expect(kerrors.IsNotFound(err)).To(BeTrue())
				})

				By("checking IPs do exist for the next iteration", func() {
					for _, svcName := range ipSvcNames {
						_, err := cpi.Kube().Services().Get(svcName, kmetav1.GetOptions{})
						Expect(err).ToNot(HaveOccurred())
					}
				})
			}

			By("cleaning up IPs", func() {
				for _, svcName := range ipSvcNames {
					err := cpi.Kube().Services().Delete(svcName, kmetav1.NewDeleteOptions(0))
					Expect(err).ToNot(HaveOccurred())
				}
			})
		})
	})
})
