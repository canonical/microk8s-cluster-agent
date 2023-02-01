package k8sinit

import (
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// Launcher is used to apply launch configurations to the MicroK8s cluster.
type Launcher struct {
	snap    snap.Snap
	preInit bool
}

// NewLauncher creates a new launcher instance.
// preInit is true when applying the configuration prior to any of the services running.
func NewLauncher(s snap.Snap, preInit bool) *Launcher {
	return &Launcher{snap: s, preInit: preInit}
}
