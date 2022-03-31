package v2

import (
	"context"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

// ListControlPlaneNodeIPsFunc returns a list of the known control plane nodes of a MicroK8s cluster.
type ListControlPlaneNodeIPsFunc func(ctx context.Context, _ snap.Snap) ([]string, error)
