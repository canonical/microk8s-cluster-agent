package k8sinit

import (
	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// Launcher is used to apply launch configurations to the MicroK8s cluster.
type Launcher struct {
	snap snap.Snap
}

// NewLauncher creates a new launcher instance.
func NewLauncher(s snap.Snap) *Launcher {
	return &Launcher{snap: s}
}
