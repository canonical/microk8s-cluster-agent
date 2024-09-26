package v2

import (
	"context"
	"fmt"
	"net/http"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

// RemoveRequest represents a request to remove a node from the dqlite cluster.
type RemoveRequest struct {
	// HostPort is the address of the node to remove.
	HostPort string
}

// Remove implements the "POST /v2/remove" endpoint and removes a node from the dqlite cluster.
func (a *API) Remove(ctx context.Context, req RemoveRequest) (int, error) {
	if err := snaputil.RemoveNodeFromDqlite(ctx, a.Snap, req.HostPort); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to remove node from dqlite: %w", err)
	}

	return http.StatusOK, nil
}
