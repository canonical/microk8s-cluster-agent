package k8sinit

import (
	"context"
	"fmt"
)

func (l *Launcher) Apply(ctx context.Context, c *Configuration) error {
	switch {
	case c == nil:
		return nil
	case c.Version > MaximumConfigFileVersionSupported:
		return fmt.Errorf("config file version is %v but the maximum version supported is %v", c.Version, MaximumConfigFileVersionSupported)
	case c.Version < MinimumConfigFileVersionRequired:
		return fmt.Errorf("config file version is %v but the minimum version required is %v", c.Version, MinimumConfigFileVersionRequired)
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
