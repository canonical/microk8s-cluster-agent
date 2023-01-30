package k8sinit

import (
	"context"
	"fmt"
)

// Apply applies a MicroK8s launch configuration to the local MicroK8s node.
func (l *Launcher) Apply(ctx context.Context, c *Configuration) error {
	if c == nil {
		return nil
	}

	for _, addon := range c.Addons {
		if err := l.applyAddon(ctx, addon); err != nil {
			return fmt.Errorf("failed to apply addon %q: %w", addon.Name, err)
		}
	}

	return nil
}

func (l *Launcher) applyAddon(ctx context.Context, c AddonConfiguration) error {
	f := l.snap.EnableAddon
	if !c.Enable {
		f = l.snap.DisableAddon
	}

	if err := f(ctx, c.Name, c.Arguments...); err != nil {
		return fmt.Errorf("addon operation failed: %w", err)
	}

	return nil
}
