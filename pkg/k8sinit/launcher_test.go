package k8sinit

import (
	"context"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
)

func TestConfigFileVersion(t *testing.T) {
	for _, tc := range []struct {
		name      string
		version   int
		expectErr bool
	}{
		{name: "newer", version: maximumConfigFileVersionSupported + 1, expectErr: true},
		{name: "unsupported", version: minimumConfigFileVersionRequired - 1, expectErr: true},
		{name: "latest", version: maximumConfigFileVersionSupported, expectErr: false},
		{name: "oldest", version: maximumConfigFileVersionSupported, expectErr: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg := Configuration{
				Version: tc.version,
			}
			err := NewLauncher(&mock.Snap{}).Apply(context.Background(), &cfg)
			switch {
			case err == nil && tc.expectErr:
				t.Fatal("expected an error but did not receive any")
			case err != nil && !tc.expectErr:
				t.Fatalf("did not expect an error but received %q", err)
			}
		})
	}
}

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
				{Name: "dns", Enable: true},
				{Name: "mayastor", Enable: true, Arguments: []string{"--default-pool-size", "20GB"}},
				{Name: "registry", Enable: false},
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
			if err := l.Apply(context.Background(), &Configuration{Version: minimumConfigFileVersionRequired, Addons: tc.addons}); err != nil {
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
