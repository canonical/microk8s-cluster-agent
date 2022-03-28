package v2

import "net"

// API implements the v2 API.
type API struct {
	// ListControlPlaneNodeIPs is used in v2/join to list the IP addresses of the
	// known control plane nodes.
	ListControlPlaneNodeIPs ListControlPlaneNodeIPsFunc

	// LookupIP is net.LookupIP.
	LookupIP func(string) ([]net.IP, error)
}
