package network

import (
	"fmt"
	"strings"

	bosherr "github.com/cloudfoundry/bosh-utils/errors"
	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	apiv1 "github.com/cppforlife/bosh-cpi-go/apiv1"
	kcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type ManualNetwork struct {
	cid            apiv1.VMCID
	network        apiv1.Network
	ingress        []Ingress
	servicesClient kcorev1.ServiceInterface

	logTag string
	logger boshlog.Logger
}

func NewManualNetwork(cid apiv1.VMCID, network apiv1.Network,
	ingress []Ingress, servicesClient kcorev1.ServiceInterface,
	logger boshlog.Logger) ManualNetwork {

	return ManualNetwork{cid, network,
		ingress, servicesClient,
		"vm.networking.ManualNetwork", logger}
}

func (n ManualNetwork) Create() error {
	ports := Ingresses(n.ingress).ToPorts()

	if len(ports) == 0 {
		// todo better error?
		return bosherr.Errorf("Expected to find at least one ingress" +
			" port configuration when using manual networks")
	}

	clusterIP := NewClusterIP(n.network.IP(), n.servicesClient, n.logger)

	return clusterIP.Attach(n.cid, ports)
}

func (n ManualNetwork) ToBashCmd() string {
	manualIP := n.network.IP()

	lines := []string{
		`dynamic_net=$(ip addr show eth0 | grep "inet\b" | awk '{print $2}')`,
		`dynamic_ip=$(echo $dynamic_net | cut -d/ -f1)`,
		// by keeping at least one ip on eth0, default route is preserved
		fmt.Sprintf("ip addr add %s/32 dev eth0", manualIP),
		// move private primary ip to be last in the list
		`ip addr del $dynamic_net dev eth0`,
		`ip addr add $dynamic_net dev eth0`,
		// rewrite packets to match to listen address (nc -l ip)
		fmt.Sprintf("iptables -t nat -A PREROUTING -d $dynamic_ip -j DNAT --to-destination %s", manualIP),
		fmt.Sprintf("iptables -t nat -A POSTROUTING -s %s         -j SNAT --to-source $dynamic_ip", manualIP),
	}

	return strings.Join(lines, " && ") // todo how to group commands?
}
