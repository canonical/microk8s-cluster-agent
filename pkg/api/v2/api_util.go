package v2

import (
	"fmt"
	"log"
	"net"
)

// findMatchingBindAddress attempts to find the bind address for dqlite from the 'host:port' of the join request.
// in case of system errors, the request host is returned, to preserve backwards-compatibility.
func (a *API) findMatchingBindAddress(hostPort string) (string, error) {
	requestHost, _, _ := net.SplitHostPort(hostPort)

	addrs, err := a.InterfaceAddrs()
	if err != nil {
		log.Printf("[WARNING] failed to retrieve host addresses: %v", err)
		log.Printf("[WARNING] will attempt to use %v as dqlite bind address", requestHost)
		return requestHost, nil
	}

nextAddr:
	for _, addr := range addrs {
		ip, subnet, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue nextAddr
		}

		// FIXME(neoaggelos): handle the case where a virtual IP is used for the microk8s join command.
		// in that scenario, the machine will have two IP addresses like this:
		//     10.0.0.11/16             <-- IP address of the machine, subnet is 10.0.0.0/16
		//     10.0.0.100/32            <-- Virtual IP, /32 address but contained in 10.0.0.0/16
		//
		// We should detect whether a virtual IP is used (in this case 10.0.0.100/32) and use the
		// respective IP address instead (in this case 10.0.0.11/16).
		//
		// One way to do this is:
		// - for all addresses, use subnet.Mask.Size() and check whether the subnet mask is all ones (e.g. /32 for IPv4, /128 for IPv6)
		// - if subnet mask is not all ones, check whether subnet.Contains(requestHost)
		//
		// If the above conditions hold, use ip.String() instead of the requestHost (needs adjustments in the condition above)
		_ = subnet

		if ip.String() == requestHost {
			// this is a known IP address, accept it
			return requestHost, nil
		}
	}

	// no host address matched
	return "", fmt.Errorf("address %v was not found in any host interface. refuse to update dqlite bind address to %v as it would break the cluster", requestHost, requestHost)
}
