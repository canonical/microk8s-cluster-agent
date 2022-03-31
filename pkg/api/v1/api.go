package v1

import (
	"net"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// API implements the v1 API.
type API struct {
	// Snap interacts with the MicroK8s snap.
	Snap snap.Snap

	// LookupIP is net.LookupIP.
	LookupIP func(string) ([]net.IP, error)
}
