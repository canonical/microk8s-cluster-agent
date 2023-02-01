package k8sinit

import (
	"context"
	"fmt"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

// Apply applies a multi-part configuration to the local MicroK8s node.
func (l *Launcher) Apply(ctx context.Context, c MultiPartConfiguration) error {
	for idx, part := range c.Parts {
		if err := l.applyPart(ctx, part); err != nil {
			return fmt.Errorf("failed to apply config part %d: %w", idx, err)
		}
	}
	return nil
}

// applyPart applies a MicroK8s launch configuration to the local MicroK8s node.
func (l *Launcher) applyPart(ctx context.Context, c *Configuration) error {
	if c == nil {
		return nil
	}

	if !l.preInit {
		l.reconcileAddons(ctx, c.Addons)
	}

	if err := l.reconcileKubeletArgs(ctx, c.ExtraKubeletArgs); err != nil {
		return fmt.Errorf("failed to configure extra kubelet args: %w", err)
	}

	return nil
}

func (l *Launcher) reconcileAddons(ctx context.Context, addons []AddonConfiguration) error {
	for _, addon := range addons {
		if addon.Disable {
			if err := l.snap.DisableAddon(ctx, addon.Name, addon.Arguments...); err != nil {
				return fmt.Errorf("failed to disable addon %q: %w", addon.Name, err)
			}
		} else if err := l.snap.EnableAddon(ctx, addon.Name, addon.Arguments...); err != nil {
			return fmt.Errorf("failed to enable addon %q: %w", addon.Name, err)
		}
	}
	return nil
}

func (l *Launcher) reconcileKubeletArgs(ctx context.Context, args map[string]*string) error {
	if len(args) == 0 {
		return nil
	}
	updateArgs := map[string]string{}
	deleteArgs := []string{}

	for key, valptr := range args {
		if valptr == nil {
			deleteArgs = append(deleteArgs, key)
		} else {
			updateArgs[key] = *valptr
		}
	}

	changed, err := snaputil.UpdateServiceArguments(l.snap, "kubelet", []map[string]string{updateArgs}, deleteArgs)
	if err != nil {
		return fmt.Errorf("failed to update service arguments: %w", err)
	}

	// TODO(neoaggelos): restart services should be deferred until the very end of the function
	if changed && !l.preInit {
		if err := l.snap.RestartService(ctx, "kubelet"); err != nil {
			return fmt.Errorf("failed to restart kubelet service after updating arguments: %w", err)
		}
	}

	return nil
}
