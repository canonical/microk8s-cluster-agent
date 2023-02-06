package k8sinit

import (
	"context"
	"fmt"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

type launcherScope struct {
	launcher *Launcher

	needsKubeliteRestart bool
}

// Apply applies a multi-part configuration to the local MicroK8s node.
func (l *Launcher) Apply(ctx context.Context, c MultiPartConfiguration) error {
	s := &launcherScope{
		launcher: l,
	}
	for idx, part := range c.Parts {
		if err := s.applyPart(ctx, part); err != nil {
			return fmt.Errorf("failed to apply config part %d: %w", idx, err)
		}
	}
	if !s.launcher.preInit && s.needsKubeliteRestart {
		if err := s.launcher.snap.RestartService(ctx, "kubelite"); err != nil {
			return fmt.Errorf("failed to restart kubelite to apply configuration: %w", err)
		}
	}
	return nil
}

// applyPart applies a MicroK8s launch configuration to the local MicroK8s node.
func (s *launcherScope) applyPart(ctx context.Context, c *Configuration) error {
	if c == nil {
		return nil
	}

	if !s.launcher.preInit {
		if err := s.reconcileAddons(ctx, c.Addons); err != nil {
			return fmt.Errorf("failed to reconcile addons: %w", err)
		}
	}

	for svc, args := range map[string]map[string]*string{
		"kube-apiserver":          c.ExtraKubeAPIServerArgs,
		"kubelet":                 c.ExtraKubeletArgs,
		"kube-proxy":              c.ExtraKubeProxyArgs,
		"kube-controller-manager": c.ExtraKubeControllerManagerArgs,
		"kube-scheduler":          c.ExtraKubeSchedulerArgs,
	} {
		if err := s.reconcileServiceArgs(ctx, svc, args); err != nil {
			return fmt.Errorf("failed to reconcile %q service flags: %w", svc, err)
		}
	}

	if err := s.reconcileExtraSANs(c.ExtraSANs); err != nil {
		return fmt.Errorf("failed to configure SANs for apiserver: %w", err)
	}

	return nil
}

func (s *launcherScope) reconcileAddons(ctx context.Context, addons []AddonConfiguration) error {
	for _, addon := range addons {
		if addon.Disable {
			if err := s.launcher.snap.DisableAddon(ctx, addon.Name, addon.Arguments...); err != nil {
				return fmt.Errorf("failed to disable addon %q: %w", addon.Name, err)
			}
		} else if err := s.launcher.snap.EnableAddon(ctx, addon.Name, addon.Arguments...); err != nil {
			return fmt.Errorf("failed to enable addon %q: %w", addon.Name, err)
		}
	}
	return nil
}

func (s *launcherScope) reconcileServiceArgs(ctx context.Context, service string, args map[string]*string) error {
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

	changed, err := snaputil.UpdateServiceArguments(s.launcher.snap, service, []map[string]string{updateArgs}, deleteArgs)
	if err != nil {
		return fmt.Errorf("failed to update service arguments: %w", err)
	}

	if changed {
		s.needsKubeliteRestart = true
	}

	return nil
}

func (s *launcherScope) reconcileExtraSANs(extraSANs []string) error {
	csr, err := util.GenerateCSRConf(extraSANs)
	if err != nil {
		return fmt.Errorf("failed to generate csr configuration: %w", err)
	}
	if err := s.launcher.snap.WriteCSRConfig(csr); err != nil {
		return fmt.Errorf("failed to write csr configuration: %w", err)
	}
	return nil
}
