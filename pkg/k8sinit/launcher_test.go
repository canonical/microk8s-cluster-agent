package k8sinit

import (
	"context"
	"fmt"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	. "github.com/onsi/gomega"
)

func TestAddons(t *testing.T) {
	for _, tc := range []struct {
		name                string
		addons              []AddonConfiguration
		expectEnableAddons  []string
		expectDisableAddons []string
	}{
		{
			name: "simple",
			addons: []AddonConfiguration{
				{Name: "dns", Disable: false},
				{Name: "mayastor", Disable: false, Arguments: []string{"--default-pool-size", "20GB"}},
				{Name: "registry", Disable: true},
			},
			expectEnableAddons: []string{
				"dns",
				"mayastor --default-pool-size 20GB",
			},
			expectDisableAddons: []string{
				"registry",
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, preInit := range []bool{false, true} {
				t.Run(fmt.Sprintf("preInit=%v", preInit), func(t *testing.T) {
					s := &mock.Snap{}

					l := NewLauncher(s, preInit)
					c := MultiPartConfiguration{[]*Configuration{
						{Version: minimumConfigFileVersionRequired.String(), Addons: tc.addons},
					}}
					g := NewWithT(t)
					err := l.Apply(context.Background(), c)
					g.Expect(err).To(BeNil())

					if !preInit {
						g.Expect(s.EnableAddonCalledWith).To(Equal(tc.expectEnableAddons))
						g.Expect(s.DisableAddonCalledWith).To(Equal(tc.expectDisableAddons))
					} else {
						g.Expect(s.EnableAddonCalledWith).To(BeEmpty())
						g.Expect(s.DisableAddonCalledWith).To(BeEmpty())
					}
				})
			}
		})
	}
}

func TestAddonRepositories(t *testing.T) {
	for _, tc := range []struct {
		name        string
		repos       []AddonRepositoryConfiguration
		expectRepos map[string]mock.AddonRepository
	}{
		{
			name: "simple",
			repos: []AddonRepositoryConfiguration{
				{Name: "core", URL: "https://github.com/canonical/microk8s-core-addons"},
				{Name: "community", URL: "https://github.com/canonical/microk8s-core-addons", Reference: "custom-branch"},
			},
			expectRepos: map[string]mock.AddonRepository{
				"core":      {URL: "https://github.com/canonical/microk8s-core-addons", Force: true},
				"community": {URL: "https://github.com/canonical/microk8s-core-addons", Reference: "custom-branch", Force: true},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, preInit := range []bool{false, true} {
				t.Run(fmt.Sprintf("preInit=%v", preInit), func(t *testing.T) {
					s := &mock.Snap{}

					l := NewLauncher(s, preInit)
					c := MultiPartConfiguration{[]*Configuration{
						{Version: minimumConfigFileVersionRequired.String(), AddonRepositories: tc.repos},
					}}
					g := NewWithT(t)
					err := l.Apply(context.Background(), c)
					g.Expect(err).To(BeNil())

					if !preInit {
						g.Expect(s.AddonRepositories).To(Equal(tc.expectRepos))
					} else {
						g.Expect(s.AddonRepositories).To(BeEmpty())
					}
				})
			}
		})
	}
}

func TestContainerdRegistryConfigs(t *testing.T) {
	s := &mock.Snap{}

	l := NewLauncher(s, false)
	c := MultiPartConfiguration{[]*Configuration{
		{
			Version: minimumConfigFileVersionRequired.String(),
			ContainerdRegistryConfigs: map[string]string{
				"docker.io": `server = "http://dockerhub.mirror:32000"`,
				"quay.io":   `server = "http://quay.mirror:32000"`,
			},
		},
	}}
	g := NewWithT(t)

	err := l.Apply(context.Background(), c)
	g.Expect(err).To(BeNil())

	g.Expect(s.ContainerdRegistryConfigs["docker.io"]).To(Equal(`server = "http://dockerhub.mirror:32000"`))
	g.Expect(s.ContainerdRegistryConfigs["quay.io"]).To(Equal(`server = "http://quay.mirror:32000"`))

	g.Expect(s.RestartServiceCalledWith).To(BeEmpty())
}

func TestComponentConfiguration(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		setConfig            func(*Configuration)
		expectServiceArgs    map[string][]string
		expectServiceRestart []string
	}{
		{
			name: "kubernetes",
			setConfig: func(c *Configuration) {
				c.ExtraKubeletArgs = map[string]*string{
					"--Kubelet-arg": &[]string{"value"}[0],
				}
				c.ExtraKubeAPIServerArgs = map[string]*string{
					"--KubeAPIServer-arg": &[]string{"value"}[0],
				}
				c.ExtraKubeProxyArgs = map[string]*string{
					"--KubeProxy-arg": &[]string{"value"}[0],
				}
				c.ExtraKubeControllerManagerArgs = map[string]*string{
					"--KubeControllerManager-arg": &[]string{"value"}[0],
				}
				c.ExtraKubeSchedulerArgs = map[string]*string{
					"--KubeScheduler-arg": &[]string{"value"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"kubelet":                 {"--Kubelet-arg=value\n"},
				"kube-apiserver":          {"--KubeAPIServer-arg=value\n"},
				"kube-proxy":              {"--KubeProxy-arg=value\n"},
				"kube-controller-manager": {"--KubeControllerManager-arg=value\n"},
				"kube-scheduler":          {"--KubeScheduler-arg=value\n"},
			},
			expectServiceRestart: []string{"kubelite"},
		},
		{
			name: "kubernetes-env",
			setConfig: func(c *Configuration) {
				c.ExtraKubeliteEnv = map[string]*string{
					"GOFIPS": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"kubelite-env": {"GOFIPS=1\n"},
			},
			expectServiceRestart: []string{"kubelite"},
		},
		{
			name: "containerd",
			setConfig: func(c *Configuration) {
				c.ExtraContainerdArgs = map[string]*string{
					"-l": &[]string{"debug"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"containerd": {"-l=debug\n"},
			},
			expectServiceRestart: []string{"containerd"},
		},
		{
			name: "containerd-env",
			setConfig: func(c *Configuration) {
				c.ExtraContainerdEnv = map[string]*string{
					"http_proxy":  &[]string{"http://squid.internal:3128"}[0],
					"https_proxy": &[]string{"http://squid.internal:3128"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"containerd-env": {"http_proxy=http://squid.internal:3128\n", "https_proxy=http://squid.internal:3128\n"},
			},
			expectServiceRestart: []string{"containerd"},
		},
		{
			name: "dqlite",
			setConfig: func(c *Configuration) {
				c.ExtraDqliteArgs = map[string]*string{
					"--disk-mode": &[]string{"true"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"k8s-dqlite": {"--disk-mode=true\n"},
			},
			expectServiceRestart: []string{"k8s-dqlite"},
		},
		{
			name: "dqlite-env",
			setConfig: func(c *Configuration) {
				c.ExtraDqliteEnv = map[string]*string{
					"LIBRAFT_TRACE":   &[]string{"1"}[0],
					"LIBDQLITE_TRACE": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"k8s-dqlite-env": {"LIBRAFT_TRACE=1\n", "LIBDQLITE_TRACE=1\n"},
			},
			expectServiceRestart: []string{"k8s-dqlite"},
		},
		{
			name: "cluster-agent",
			setConfig: func(c *Configuration) {
				c.ExtraMicroK8sClusterAgentArgs = map[string]*string{
					"--bind": &[]string{"10.0.0.10:25000"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"cluster-agent": {"--bind=10.0.0.10:25000\n"},
			},
			expectServiceRestart: []string{"cluster-agent"},
		},
		{
			name: "cluster-agent-env",
			setConfig: func(c *Configuration) {
				c.ExtraMicroK8sClusterAgentEnv = map[string]*string{
					"GOFIPS": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"cluster-agent-env": {"GOFIPS=1\n"},
			},
			expectServiceRestart: []string{"cluster-agent"},
		},
		{
			name: "apiserver-proxy",
			setConfig: func(c *Configuration) {
				c.ExtraMicroK8sAPIServerProxyArgs = map[string]*string{
					"--refresh-interval": &[]string{"0s"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"apiserver-proxy": {"--refresh-interval=0s\n"},
			},
			expectServiceRestart: []string{"apiserver-proxy"},
		},
		{
			name: "apiserver-proxy-env",
			setConfig: func(c *Configuration) {
				c.ExtraMicroK8sAPIServerProxyEnv = map[string]*string{
					"GOFIPS": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"apiserver-proxy-env": {"GOFIPS=1\n"},
			},
			expectServiceRestart: []string{"apiserver-proxy"},
		},
		{
			name: "etcd",
			setConfig: func(c *Configuration) {
				c.ExtraEtcdArgs = map[string]*string{
					"--enable-v2": &[]string{"false"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"etcd": {"--enable-v2=false\n"},
			},
			expectServiceRestart: []string{"etcd"},
		},
		{
			name: "etcd-env",
			setConfig: func(c *Configuration) {
				c.ExtraEtcdEnv = map[string]*string{
					"GOFIPS": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"etcd-env": {"GOFIPS=1\n"},
			},
			expectServiceRestart: []string{"etcd"},
		},
		{
			name: "flanneld",
			setConfig: func(c *Configuration) {
				c.ExtraFlanneldArgs = map[string]*string{
					"--ip-masq": &[]string{"false"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"flanneld": {"--ip-masq=false\n"},
			},
			expectServiceRestart: []string{"flanneld"},
		},
		{
			name: "flanneld-env",
			setConfig: func(c *Configuration) {
				c.ExtraFlanneldEnv = map[string]*string{
					"GOFIPS": &[]string{"1"}[0],
				}
			},
			expectServiceArgs: map[string][]string{
				"flanneld-env": {"GOFIPS=1\n"},
			},
			expectServiceRestart: []string{"flanneld"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, preInit := range []bool{false, true} {
				t.Run(fmt.Sprintf("preInit=%v", preInit), func(t *testing.T) {

					s := &mock.Snap{}

					l := NewLauncher(s, preInit)
					c := MultiPartConfiguration{[]*Configuration{{
						Version: minimumConfigFileVersionRequired.String(),
					}}}

					tc.setConfig(c.Parts[0])

					g := NewWithT(t)

					err := l.Apply(context.Background(), c)
					g.Expect(err).To(BeNil())

					for svc, fragments := range tc.expectServiceArgs {
						for _, fragment := range fragments {
							g.Expect(s.ServiceArguments[svc]).To(ContainSubstring(fragment))
						}
					}

					if preInit {
						g.Expect(s.RestartServiceCalledWith).To(BeEmpty())
					} else {
						g.Expect(s.RestartServiceCalledWith).To(ConsistOf(tc.expectServiceRestart))
					}
				})
			}
		})
	}
}
