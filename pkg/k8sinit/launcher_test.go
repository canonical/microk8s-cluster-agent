package k8sinit

import (
	"context"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
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

			l := NewLauncher(s)
			if err := l.Apply(context.Background(), &Configuration{Version: minimumConfigFileVersionRequired.String(), Addons: tc.addons}); err != nil {
				t.Fatalf("expected no error when applying configuration but got %q instead", err)
			}

			if !reflect.DeepEqual(s.EnableAddonCalledWith, tc.expectEnableAddons) {
				t.Fatalf("expected enabled addons %v but got %v instead", tc.expectEnableAddons, s.EnableAddonCalledWith)
			}
			if !reflect.DeepEqual(s.DisableAddonCalledWith, tc.expectDisableAddons) {
				t.Fatalf("expected disabled addons %v but got %v instead", tc.expectDisableAddons, s.DisableAddonCalledWith)
			}
		})
	}
}
