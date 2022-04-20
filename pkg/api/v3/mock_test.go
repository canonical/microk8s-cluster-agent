package v3_test

import (
	"context"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

func mockListControlPlaneNodes(nodes ...string) func(context.Context, snap.Snap) ([]string, error) {
	return func(context.Context, snap.Snap) ([]string, error) {
		return nodes, nil
	}
}
