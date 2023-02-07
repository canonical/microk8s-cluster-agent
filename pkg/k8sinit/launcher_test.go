package k8sinit

import (
	"context"
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
			s := &mock.Snap{}

			l := NewLauncher(s, false)
			c := MultiPartConfiguration{[]*Configuration{
				{Version: minimumConfigFileVersionRequired.String(), Addons: tc.addons},
			}}
			g := NewWithT(t)
			err := l.Apply(context.Background(), c)
			g.Expect(err).To(BeNil())

			g.Expect(s.EnableAddonCalledWith).To(Equal(tc.expectEnableAddons))
			g.Expect(s.DisableAddonCalledWith).To(Equal(tc.expectDisableAddons))
		})
	}
}

func TestExtraServiceArguments(t *testing.T) {
	s := &mock.Snap{}

	l := NewLauncher(s, false)
	c := MultiPartConfiguration{[]*Configuration{
		{
			Version: minimumConfigFileVersionRequired.String(),
			ExtraKubeletArgs: map[string]*string{
				"--Kubelet-arg": &[]string{"value"}[0],
			},
			ExtraKubeAPIServerArgs: map[string]*string{
				"--KubeAPIServer-arg": &[]string{"value"}[0],
			},
			ExtraKubeProxyArgs: map[string]*string{
				"--KubeProxy-arg": &[]string{"value"}[0],
			},
			ExtraKubeControllerManagerArgs: map[string]*string{
				"--KubeControllerManager-arg": &[]string{"value"}[0],
			},
			ExtraKubeSchedulerArgs: map[string]*string{
				"--KubeScheduler-arg": &[]string{"value"}[0],
			},
		},
	}}

	g := NewWithT(t)
	err := l.Apply(context.Background(), c)
	g.Expect(err).To(BeNil())

	g.Expect(s.WriteServiceArgumentsCalled).To(BeTrue())

	g.Expect(s.ServiceArguments["kubelet"]).To(ContainSubstring("--Kubelet-arg=value"))
	g.Expect(s.ServiceArguments["kube-apiserver"]).To(ContainSubstring("--KubeAPIServer-arg=value"))
	g.Expect(s.ServiceArguments["kube-proxy"]).To(ContainSubstring("--KubeProxy-arg=value"))
	g.Expect(s.ServiceArguments["kube-controller-manager"]).To(ContainSubstring("--KubeControllerManager-arg=value"))
	g.Expect(s.ServiceArguments["kube-scheduler"]).To(ContainSubstring("--KubeScheduler-arg=value"))

	g.Expect(s.RestartServiceCalledWith).To(ContainElement("kubelite"))
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
}
