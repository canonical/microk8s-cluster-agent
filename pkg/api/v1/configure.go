package v1

import (
	"context"
	"fmt"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

// ConfigureServiceRequest is a configuration request for MicroK8s.
type ConfigureServiceRequest struct {
	// Name is the service name.
	Name string `json:"name"`
	// UpdateArguments is a map of arguments to be updated.
	UpdateArguments []map[string]string `json:"arguments_update"`
	// RemoveArguments is a list of arguments to remove.
	RemoveArguments []string `json:"arguments_remove"`
	// Restart defines whether the service should be restarted.
	Restart RestartServiceField `json:"restart"`
}

// ConfigureAddonRequest is a configuration request for a MicroK8s addon.
type ConfigureAddonRequest struct {
	// Name is the name of the addon.
	Name string `json:"name"`
	// Enable is true if we want to enable the addon.
	Enable bool `json:"enable"`
	// Disable is true if we want to disable the addon.
	Disable bool `json:"disable"`
}

// ConfigureRequest is the request message for the v1/configure endpoint.
type ConfigureRequest struct {
	// CallbackToken is the callback token used to authenticate the request.
	CallbackToken string `json:"callback"`

	// ConfigureServices is a list of configuration updates for the MicroK8s services.
	ConfigureServices []ConfigureServiceRequest `json:"service"`

	// ConfigureAddons is a list of addons to manage
	ConfigureAddons []ConfigureAddonRequest `json:"addon"`
}

// Configure implements "POST /CLUSTER_API_V1/configure".
func (a *API) Configure(ctx context.Context, req ConfigureRequest) error {
	if !a.Snap.ConsumeSelfCallbackToken(req.CallbackToken) {
		return fmt.Errorf("invalid token")
	}
	for _, service := range req.ConfigureServices {
		if err := snaputil.UpdateServiceArguments(a.Snap, service.Name, service.UpdateArguments, service.RemoveArguments); err != nil {
			return fmt.Errorf("failed to update arguments of service %q: %w", service.Name, err)
		}
		if service.Restart {
			if err := a.Snap.RestartService(ctx, service.Name); err != nil {
				return fmt.Errorf("failed to restart service %q: %w", service.Name, err)
			}
		}
	}

	for _, addon := range req.ConfigureAddons {
		switch {
		case addon.Enable:
			if err := a.Snap.EnableAddon(ctx, addon.Name); err != nil {
				return fmt.Errorf("failed to enable addon %q: %w", addon.Name, err)
			}
		case addon.Disable:
			if err := a.Snap.DisableAddon(ctx, addon.Name); err != nil {
				return fmt.Errorf("failed to disable addon %q: %w", addon.Name, err)
			}
		}
	}
	return nil
}
