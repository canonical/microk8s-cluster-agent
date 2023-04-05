package v2

import (
	"net"
	"sync"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// API implements the v2 API.
type API struct {
	// Snap interacts with the MicroK8s snap.
	Snap snap.Snap

	// ListControlPlaneNodeIPs is used in v2/join to list the IP addresses of the
	// known control plane nodes.
	ListControlPlaneNodeIPs ListControlPlaneNodeIPsFunc

	// LookupIP is net.LookupIP.
	LookupIP func(string) ([]net.IP, error)

	// InterfaceAddrs is net.InterfaceAddrs.
	InterfaceAddrs func() ([]net.Addr, error)

	// dqliteMu protects changes involving the dqlite service.
	dqliteMu sync.Mutex

	// calicoMu protects changes involving the calico CNI.
	calicoMu sync.Mutex
}
