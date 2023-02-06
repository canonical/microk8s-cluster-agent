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
			if err := l.Apply(context.Background(), c); err != nil {
				t.Fatalf("expected no error when applying configuration but got %q instead", err)
			}

			g := NewWithT(t)

			g.Expect(s.EnableAddonCalledWith).To(Equal(tc.expectEnableAddons))
			g.Expect(s.DisableAddonCalledWith).To(Equal(tc.expectDisableAddons))
		})
	}
}
	}
}
