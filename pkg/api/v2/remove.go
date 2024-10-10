package v2

import (
	"context"
	"fmt"
	"net/http"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

// RemoveFromDqliteRequest represents a request to remove a node from the dqlite cluster.
type RemoveFromDqliteRequest struct {
	// RemoveEndpoint is the endpoint of the node to remove from the dqlite cluster.
	RemoveEndpoint string `json:"remove_endpoint"`
}

// RemoveFromDqlite implements the "POST /v2/dqlite/remove" endpoint and removes a node from the dqlite cluster.
func (a *API) RemoveFromDqlite(ctx context.Context, req RemoveFromDqliteRequest, token string) (int, error) {
	isValid, err := a.Snap.IsCAPIAuthTokenValid(token)
	if err != nil {
		return http.StatusUnauthorized, fmt.Errorf("failed to validate CAPI auth token: %w", err)
	}

	if !isValid {
		return http.StatusUnauthorized, fmt.Errorf("invalid CAPI auth token %q", token)
	}

	if err := snaputil.RemoveNodeFromDqlite(ctx, a.Snap, req.RemoveEndpoint); err != nil {
		return http.StatusInternalServerError, fmt.Errorf("failed to remove node from dqlite: %w", err)
	}

	return http.StatusOK, nil
}
