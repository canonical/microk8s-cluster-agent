package v2

import (
	"fmt"
	"log"
	"net"
	"strings"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

// findMatchingBindAddress attempts to find the bind address for dqlite from the 'host:port' of the join request.
// in case of system errors, the request host is returned, to preserve backwards-compatibility.
func (a *API) findMatchingBindAddress(hostPort string) (string, error) {
	hostIP, _, _ := net.SplitHostPort(hostPort)

	hostNetIP := net.ParseIP(hostIP)
	if hostNetIP == nil {
		log.Printf("[WARNING] failed to parse IP address %v", hostIP)
		log.Printf("[WARNING] will attempt to use %v as dqlite bind address", hostIP)
		return hostIP, nil
	}

	addrs, err := a.InterfaceAddrs()
	if err != nil {
		log.Printf("[WARNING] failed to retrieve host addresses: %v", err)
		log.Printf("[WARNING] will attempt to use %v as dqlite bind address", hostIP)
		return hostIP, nil
	}

	var (
		isVirtualIP         bool
		matchingInterfaceIP net.IP
	)

nextAddr:
	for _, addr := range addrs {
		ip, subnet, err := net.ParseCIDR(addr.String())
		if err != nil || subnet == nil {
			log.Printf("[WARNING] failed to parse address %v: %v", addr.String(), err)
			continue nextAddr
		}

		ones, bits := subnet.Mask.Size()
		subnetHostBits := bits - ones
		if ip.Equal(hostNetIP) {
			// virtual IPs are /32 IPv4 or /128 IPv6
			isVirtualIP = subnetHostBits == 0
			if !isVirtualIP {
				return hostIP, nil
			}
		} else if subnet.Contains(hostNetIP) && subnetHostBits > 0 {
			// we found the IP address of the interface
			matchingInterfaceIP = ip
		}
	}

	if isVirtualIP {
		if matchingInterfaceIP != nil {
			return matchingInterfaceIP.String(), nil
		}

		// hostIP is most likely a virtual IP, but we were not able to find the matching IP address. return the IP address to maintain backwards-compatibility.
		return hostIP, nil
	}

	// no host address matched
	return "", fmt.Errorf("address %v was not found in any host interface. refuse to update dqlite bind address to %v as it would break the cluster", hostIP, hostIP)
}

// kubeAPIServerPrefersInternalIPForKubelet checks whether the --kubelet-preferred-address-types of kube-apiserver includes 'InternalIP' with higher preference over 'Hostname'
func (a *API) kubeAPIServerPrefersInternalIPForKubelet() bool {
	order := snaputil.GetServiceArgument(a.Snap, "kube-apiserver", "--kubelet-preferred-address-types")

	// 'Hostname' has precedence by default, argument must contain 'InternalIP'
	if ipIndex := strings.Index(order, "InternalIP"); ipIndex != -1 {
		hostnameIndex := strings.Index(order, "Hostname")
		// true if 'Hostname' is missing or 'InternalIP' is before 'Hostname'
		return hostnameIndex == -1 || ipIndex < hostnameIndex
	}

	return false
}
