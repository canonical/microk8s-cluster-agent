package v1

import (
	"context"
	"fmt"
)

// UpgradeRequest is the request message for the v1/upgrade endpoint.
type UpgradeRequest struct {
	// CallbackToken is the callback token used to authenticate the request.
	CallbackToken string `json:"callback"`
	// UpgradeName is the name of the upgrade to perform. Upgrades are listed in the `upgrade-scripts` directory,
	// for example "000-switch-to-calico".
	UpgradeName string `json:"upgrade"`
	// UpgradePhase is the current phase of the upgrade to perform. We do cluster upgrades with a 2-phase commit
	// mechanism. This can be "prepare", "commit", or "rollback".
	UpgradePhase string `json:"phase"`
}

// Upgrade implements "POST v1/upgrade".
func (a *API) Upgrade(ctx context.Context, req UpgradeRequest) error {
	if !a.Snap.ConsumeSelfCallbackToken(req.CallbackToken) {
		return fmt.Errorf("invalid token")
	}
	if err := a.Snap.RunUpgrade(ctx, req.UpgradeName, req.UpgradePhase); err != nil {
		return fmt.Errorf("failed to run upgrade %q phase %q: %w", req.UpgradeName, req.UpgradePhase, err)
	}
	return nil
}
