package k8sinit

import (
	"context"
	"fmt"
	"strings"

	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

type launcherScope struct {
	launcher *Launcher

	mustRestartServices map[string]struct{}
}

// Apply applies a multi-part configuration to the local MicroK8s node.
func (l *Launcher) Apply(ctx context.Context, c MultiPartConfiguration) error {
	s := &launcherScope{
		launcher:            l,
		mustRestartServices: make(map[string]struct{}),
	}
	for idx, part := range c.Parts {
		if err := s.applyPart(ctx, part); err != nil {
			return fmt.Errorf("failed to apply config part %d: %w", idx, err)
		}
	}
	if !s.launcher.preInit {
		for svc := range s.mustRestartServices {
			if err := s.launcher.snap.RestartService(ctx, svc); err != nil {
				return fmt.Errorf("failed to restart service %s to apply configuration: %w", svc, err)
			}
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
		if err := s.reconcileAddonRepositories(ctx, c.AddonRepositories); err != nil {
			return fmt.Errorf("failed to reconcile addon repositories: %w", err)
		}
		if err := s.reconcileAddons(ctx, c.Addons); err != nil {
			return fmt.Errorf("failed to reconcile addons: %w", err)
		}
	}

	if v := c.PersistentClusterToken; v != "" {
		if err := s.launcher.snap.AddPersistentClusterToken(v); err != nil {
			return fmt.Errorf("failed to configure persistent token: %w", err)
		}
	}

	for file, contents := range c.ExtraConfigFiles {
		if strings.Contains("/", file) {
			return fmt.Errorf("file name %q must not contain any slashes (possible path-traversal prevented)", file)
		}
		if err := s.launcher.snap.WriteServiceArguments(file, []byte(contents)); err != nil {
			return fmt.Errorf("failed to create extra config file %q: %w", file, err)
		}
	}

	for _, item := range []struct {
		configFile     string
		restartService string
		args           map[string]*string
	}{
		{configFile: "kube-apiserver", restartService: "kubelite", args: c.ExtraKubeAPIServerArgs},
		{configFile: "kubelet", restartService: "kubelite", args: c.ExtraKubeletArgs},
		{configFile: "kube-proxy", restartService: "kubelite", args: c.ExtraKubeProxyArgs},
		{configFile: "kube-controller-manager", restartService: "kubelite", args: c.ExtraKubeControllerManagerArgs},
		{configFile: "kube-scheduler", restartService: "kubelite", args: c.ExtraKubeSchedulerArgs},
		{configFile: "kubelite-env", restartService: "kubelite", args: c.ExtraKubeliteEnv},
		{configFile: "containerd", restartService: "containerd", args: c.ExtraContainerdArgs},
		{configFile: "containerd-env", restartService: "containerd", args: c.ExtraContainerdEnv},
		{configFile: "k8s-dqlite", restartService: "k8s-dqlite", args: c.ExtraDqliteArgs},
		{configFile: "k8s-dqlite-env", restartService: "k8s-dqlite", args: c.ExtraDqliteEnv},
		{configFile: "cluster-agent", restartService: "cluster-agent", args: c.ExtraMicroK8sClusterAgentArgs},
		{configFile: "cluster-agent-env", restartService: "cluster-agent", args: c.ExtraMicroK8sClusterAgentEnv},
		{configFile: "apiserver-proxy", restartService: "apiserver-proxy", args: c.ExtraMicroK8sAPIServerProxyArgs},
		{configFile: "apiserver-proxy-env", restartService: "apiserver-proxy", args: c.ExtraMicroK8sAPIServerProxyEnv},
		{configFile: "etcd", restartService: "etcd", args: c.ExtraEtcdArgs},
		{configFile: "etcd-env", restartService: "etcd", args: c.ExtraEtcdEnv},
		{configFile: "flanneld", restartService: "flanneld", args: c.ExtraFlanneldArgs},
		{configFile: "flanneld-env", restartService: "flanneld", args: c.ExtraFlanneldEnv},
	} {
		if changed, err := s.reconcileServiceArgs(ctx, item.configFile, item.args); err != nil {
			return fmt.Errorf("failed to reconcile config file %q: %w", item.configFile, err)
		} else if changed {
			s.mustRestartServices[item.restartService] = struct{}{}
		}
	}

	if err := s.reconcileExtraSANs(c.ExtraSANs); err != nil {
		return fmt.Errorf("failed to configure SANs for apiserver: %w", err)
	}

	if err := s.reconcileContainerdRegistryConfigs(c.ContainerdRegistryConfigs); err != nil {
		return fmt.Errorf("failed to reconcile containerd registry configs: %w", err)
	}

	if !s.launcher.preInit {
		if j := c.Join; j.URL != "" {
			if err := s.launcher.snap.JoinCluster(ctx, j.URL, j.Worker); err != nil {
				return fmt.Errorf("failed to join cluster: %w", err)
			}
		}
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

func (s *launcherScope) reconcileServiceArgs(ctx context.Context, configFile string, args map[string]*string) (bool, error) {
	if len(args) == 0 {
		return false, nil
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

	changed, err := snaputil.UpdateServiceArguments(s.launcher.snap, configFile, []map[string]string{updateArgs}, deleteArgs)
	if err != nil {
		return false, fmt.Errorf("failed to update arguments: %w", err)
	}
	return changed, nil
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

func (s *launcherScope) reconcileContainerdRegistryConfigs(configs map[string]string) error {
	if len(configs) == 0 {
		return nil
	}
	cfgs := make(map[string][]byte, len(configs))
	for registry, hostsToml := range configs {
		cfgs[registry] = []byte(hostsToml)
	}

	if err := s.launcher.snap.UpdateContainerdRegistryConfigs(cfgs); err != nil {
		return fmt.Errorf("failed to update containerd registry configs: %w", err)
	}
	return nil
}

func (s *launcherScope) reconcileAddonRepositories(ctx context.Context, repos []AddonRepositoryConfiguration) error {
	if len(repos) == 0 {
		return nil
	}
	for _, repo := range repos {
		if err := s.launcher.snap.AddAddonsRepository(ctx, repo.Name, repo.URL, repo.Reference, true); err != nil {
			return fmt.Errorf("failed to add repository %s: %w", repo.Name, err)
		}
	}
	return nil
}
