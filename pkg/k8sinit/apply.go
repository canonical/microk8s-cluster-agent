package k8sinit

import (
	"context"
	"fmt"
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
