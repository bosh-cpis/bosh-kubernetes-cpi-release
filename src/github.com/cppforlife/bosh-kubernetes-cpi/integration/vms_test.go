package integration_test

import (
	"encoding/json"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	kmetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/cppforlife/bosh-kubernetes-cpi/integration/testlib"
)

var _ = Describe("VMs", func() {
	Describe("create_vm", func() {
		It("can create simple vm", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_vm",
				Arguments: []interface{}{
					"agent-id",
					"scl-123/" + props.DummyImage(),
					map[string]interface{}{"instance_type": "m1.small"},
					map[string]interface{}{
						"private": map[string]interface{}{
							"type":             "dynamic",
							"dns":              []string{"8.8.8.8"},
							"default":          []string{"dns", "gateway"},
							"cloud_properties": nil,
						},
					},
					nil,
					map[string]interface{}{"env-key": "env-val"},
				},
			})
			Expect(strings.HasPrefix(resp.StringResult(), "vm-")).To(
				BeTrue(), "vm does not have nice cid prefix")

			By("checking created pod", func() {
				pod, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pod.ObjectMeta.Name).To(Equal(resp.StringResult()))
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"bosh.io/vm-cid": resp.StringResult(),
				}))
				Expect(pod.ObjectMeta.Annotations).To(BeEmpty())

				// Find generated servcice account uid
				servAcctMount := pod.Spec.Containers[0].VolumeMounts[2]
				Expect(servAcctMount.MountPath).To(Equal("/var/run/secrets/kubernetes.io/serviceaccount"))

				obj := KubeObj{pod.Spec}.Remove("schedulerName").Remove("nodeName").
					Remove("serviceAccountName").Remove("serviceAccount")

				obj.ExpectToEqual(map[string]interface{}{
					"hostname":                      "bosh-stemcell",
					"restartPolicy":                 "Never",
					"dnsPolicy":                     "ClusterFirst",
					"terminationGracePeriodSeconds": 30,
					"automountServiceAccountToken":  true, // todo should be false
					"securityContext":               map[string]interface{}{},

					"containers": []map[string]interface{}{
						{
							"securityContext": map[string]interface{}{
								"runAsUser":  0,
								"privileged": true,
							},
							"imagePullPolicy": "IfNotPresent",
							"name":            "bosh-container",
							"image":           "gcr.io/google_containers/pause-amd64:3.0",
							"command":         []string{"/bin/bash"},
							"args": []string{
								"-c",
								"umount /etc/resolv.conf && " +
									"umount /etc/hosts && " +
									"umount /etc/hostname &&  :  && " +
									"rm -rf /var/vcap/data/sys && " +
									"mkdir -p /var/vcap/data/sys && " +
									"mkdir -p /var/vcap/store && " +
									"echo '#!/bin/bash' > /var/vcap/bosh/bin/ntpdate && " +
									"exec env -i /usr/sbin/runsvdir-start",
							},
							"resources": map[string]interface{}{},
							"volumeMounts": []map[string]interface{}{
								{
									"subPath":   "warden-cpi-agent-env.json",
									"mountPath": "/var/vcap/bosh/warden-cpi-agent-env.json",
									"name":      "bosh-agent-settings",
								},
								{
									"mountPath": "/var/vcap/data",
									"name":      "bosh-ephemeral-disk",
								},
								{
									"name":      servAcctMount.Name,
									"readOnly":  true,
									"mountPath": "/var/run/secrets/kubernetes.io/serviceaccount",
								},
							},
							"terminationMessagePath":   "/dev/termination-log",
							"terminationMessagePolicy": "File",
						},
					},

					"volumes": []map[string]interface{}{
						{
							"configMap": map[string]interface{}{
								"defaultMode": 420,
								"items": []map[string]interface{}{
									{
										"path": "warden-cpi-agent-env.json",
										"key":  "instance_settings",
									},
								},
								"name": resp.StringResult(),
							},
							"name": "bosh-agent-settings",
						},
						{
							"emptyDir": map[string]interface{}{},
							"name":     "bosh-ephemeral-disk",
						},
						{
							"secret": map[string]interface{}{
								"defaultMode": 420,
								"secretName":  servAcctMount.Name,
							},
							"name": servAcctMount.Name,
						},
					},
				})
			})

			By("checking created config map", func() {
				cfgMap, err := cpi.Kube().ConfigMaps().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(cfgMap.ObjectMeta.Name).To(Equal(resp.StringResult()))
				Expect(cfgMap.ObjectMeta.Labels).To(Equal(map[string]string{
					"bosh.io/vm-cid": resp.StringResult(),
				}))
				Expect(cfgMap.ObjectMeta.Annotations).To(BeEmpty())

				var data interface{}

				err = json.Unmarshal([]byte(cfgMap.Data["instance_settings"]), &data)
				Expect(err).ToNot(HaveOccurred())

				betterEnv := KubeObj{data}.Remove("vm").
					Remove("mbus").Remove("ntp").Remove("blobstore")

				betterEnv.ExpectToEqual(map[string]interface{}{
					"agent_id": "agent-id",
					"env":      map[string]interface{}{"env-key": "env-val"},

					"disks": map[string]interface{}{
						"ephemeral":  nil,
						"persistent": nil,
						"system":     nil,
					},

					"networks": map[string]interface{}{
						"private": map[string]interface{}{
							"gateway":       "",
							"ip":            "",
							"mac":           "",
							"netmask":       "",
							"preconfigured": true, // all networks are preconfigured
							"type":          "dynamic",
							"default":       []string{"dns", "gateway"},
							"dns":           []string{"8.8.8.8"},
						},
					},
				})
			})

			By("clean up", func() {
				cpi.Exec(testlib.CPIRequest{
					Method:    "delete_vm",
					Arguments: []interface{}{resp.StringResult()},
				})

				_, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).To(HaveOccurred())
				Expect(kerrors.IsNotFound(err)).To(BeTrue())

				_, err = cpi.Kube().ConfigMaps().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).To(HaveOccurred())
				Expect(kerrors.IsNotFound(err)).To(BeTrue())

				listOpts := kmetav1.ListOptions{LabelSelector: "bosh.io/vm-cid=" + resp.StringResult()}

				svcs, err := cpi.Kube().Services().List(listOpts)
				Expect(err).ToNot(HaveOccurred())
				Expect(svcs.Items).To(BeEmpty())
			})
		})

		It("can create simple vm with custom configurations", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_vm",
				Arguments: []interface{}{
					"agent-id",
					"scl-123/" + props.DummyImage(),
					map[string]interface{}{
						"labels": map[string]string{
							"lbl1_name": "lbl1_val",
							"lbl2_name": "lbl2_val",
						},
						"annotations": map[string]string{
							"ann1_name": "ann1_val",
							"ann2_name": "ann2_val",
						},
						"resources": map[string]interface{}{
							"requests": map[string]interface{}{
								"cpu":            1,
								"memory":         "1000Mi",
								"bosh.io/custom": "100",
							},
						},
					},
					map[string]interface{}{
						"private": map[string]interface{}{
							"type":             "dynamic",
							"dns":              []string{"8.8.8.8"},
							"default":          []string{"dns", "gateway"},
							"cloud_properties": nil,
						},
					},
					nil,
					map[string]interface{}{"env-key": "env-val"},
				},
			})

			By("checking created pod", func() {
				pod, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pod.ObjectMeta.Name).To(Equal(resp.StringResult()))
				Expect(pod.ObjectMeta.Labels).To(Equal(map[string]string{
					"bosh.io/vm-cid": resp.StringResult(),
					"lbl1_name":      "lbl1_val",
					"lbl2_name":      "lbl2_val",
				}))
				Expect(pod.ObjectMeta.Annotations).To(Equal(map[string]string{
					"ann1_name": "ann1_val",
					"ann2_name": "ann2_val",
				}))

				KubeObj{pod.Spec.Containers[0]}.Find("resources").ExpectToEqual(map[string]interface{}{
					"requests": map[string]interface{}{
						"bosh.io/custom": "100",
						"cpu":            "1",
						"memory":         "1000Mi",
					},
				})
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{resp.StringResult()},
			})
		})
	})

	Describe("delete_vm", func() {
		It("can delete non-existent vm", func() {
			cpi := NewCPI()

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{"fake-does-not-exist"},
			})
		})
	})

	Describe("set_vm_metadata", func() {
		It("succeeds adding annotations", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_vm",
				Arguments: []interface{}{
					"agent-id",
					"scl-123/" + props.DummyImage(),
					map[string]interface{}{
						"annotations": map[string]string{
							"ann1_name": "ann1_val",
							"ann2_name": "ann2_val",
						},
					},
					props.RandomNetwork(),
					nil,
					nil,
				},
			})

			cpi.Exec(testlib.CPIRequest{
				Method: "set_vm_metadata",
				Arguments: []interface{}{
					resp.StringResult(),
					map[string]interface{}{
						"meta-ann1_name": "meta-ann1_val",
						"meta-ann2_name": "meta-ann2_val",
					},
				},
			})

			By("checking pod", func() {
				pod, err := cpi.Kube().Pods().Get(resp.StringResult(), kmetav1.GetOptions{})
				Expect(err).ToNot(HaveOccurred())

				Expect(pod.ObjectMeta.Annotations).To(Equal(map[string]string{
					"ann1_name":              "ann1_val",
					"ann2_name":              "ann2_val",
					"bosh.io/meta-ann1_name": "meta-ann1_val",
					"bosh.io/meta-ann2_name": "meta-ann2_val",
				}))
			})

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{resp.StringResult()},
			})
		})
	})

	Describe("has_vm", func() {
		It("returns true if vm exists", func() {
			props := NewCPIProps()
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "create_vm",
				Arguments: []interface{}{
					"agent-id",
					"scl-123/" + props.DummyImage(),
					map[string]interface{}{},
					props.RandomNetwork(),
					nil,
					nil,
				},
			})

			resp2 := cpi.Exec(testlib.CPIRequest{
				Method:    "has_vm",
				Arguments: []interface{}{resp.StringResult()},
			})
			Expect(resp2.BoolResult()).To(BeTrue())

			cpi.Exec(testlib.CPIRequest{
				Method:    "delete_vm",
				Arguments: []interface{}{resp.StringResult()},
			})
		})

		It("returns false if vm does not exists", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method:    "has_vm",
				Arguments: []interface{}{"fake-does-not-exist"},
			})
			Expect(resp.BoolResult()).To(BeFalse())
		})
	})

	Describe("reboot_vm", func() {
		It("returns error since it's not supported", func() {
			cpi := NewCPI()

			resp := cpi.ExecWithoutErrorCheck(testlib.CPIRequest{
				Method:    "reboot_vm",
				Arguments: []interface{}{"fake-vm-id"},
			})
			Expect(resp.HasError()).To(BeTrue())
			Expect(resp.ErrorMessage()).To(Equal("Rebooting is not supported"))
		})
	})

	Describe("calculate_vm_cloud_properties", func() {
		It("returns calculated properties", func() {
			cpi := NewCPI()

			resp := cpi.Exec(testlib.CPIRequest{
				Method: "calculate_vm_cloud_properties",
				Arguments: []interface{}{
					map[string]interface{}{
						"cpu": 2,
						"ram": 2048,
						"ephemeral_disk_size": 4096,
					},
				},
			})

			resp.ExpectResultToEqual(map[string]interface{}{
				"limits":    map[string]interface{}{"memory": "2048Mi", "cpu": 2},
				"resources": map[string]interface{}{"cpu": 2, "memory": "2048Mi"},
			})
		})
	})
})
