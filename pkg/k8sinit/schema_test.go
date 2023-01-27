package k8sinit_test

import (
	"embed"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/k8sinit"
)

//go:embed testdata
var testdata embed.FS

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name                string
		expectConfiguration *k8sinit.Configuration
		expectErr           bool
	}{
		{
			name: "full.yaml",
			expectConfiguration: &k8sinit.Configuration{
				Version: 1,
				Addons: []k8sinit.AddonConfiguration{
					{Name: "dns", Enable: true},
					{Name: "mayastor", Enable: true, Arguments: []string{"--default-pool-size", "20GB"}},
					{Name: "registry", Enable: false},
				},
			},
		},
		{
			name:      "invalid-yaml.yaml",
			expectErr: true,
		},
		{
			name:      "invalid-schema.yaml",
			expectErr: true,
		},
		{
			name: "unknown-fields.yaml",
			expectConfiguration: &k8sinit.Configuration{
				Version: 1,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b, err := testdata.ReadFile(filepath.Join("testdata", "schema", filepath.Join(tc.name)))
			if err != nil {
				panic(err)
			}

			c, err := k8sinit.ParseConfiguration(b)
			switch {
			case err != nil && !tc.expectErr:
				t.Fatalf("did not expect an error but got %q instead", err)
			case err == nil && tc.expectErr:
				t.Fatal("expected an error but did not get any")
			case err != nil && c != nil:
				t.Fatalf("expected empty configuration on error but got %#v instead", c)
			case err == nil && !reflect.DeepEqual(c, tc.expectConfiguration):
				t.Fatalf("Expected configuration %#v but parsed %#v instead", tc.expectConfiguration, c)
			}
		})
	}
}
