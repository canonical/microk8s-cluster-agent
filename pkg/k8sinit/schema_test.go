package k8sinit_test

import (
	"embed"
	"path/filepath"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/k8sinit"

	. "github.com/onsi/gomega"
)

//go:embed testdata
var testdata embed.FS

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		name                string
		expectConfiguration k8sinit.MultiPartConfiguration
		expectErr           bool
	}{
		{
			name: "full.yaml",
			expectConfiguration: k8sinit.MultiPartConfiguration{
				Parts: []*k8sinit.Configuration{{
					Version: "0.1.0",
					ExtraSANs: []string{
						"10.10.10.10",
						"microk8s.example.com",
					},
					ExtraKubeletArgs: map[string]*string{
						"--cluster-dns": &[]string{"10.152.183.10"}[0],
					},
					ExtraKubeAPIServerArgs: map[string]*string{
						"--authorization-mode": &[]string{"RBAC,Node"}[0],
						"--event-ttl":          nil,
					},
					ExtraKubeProxyArgs: map[string]*string{
						"--cluster-cidr": &[]string{"10.1.0.0/16"}[0],
					},
					ExtraKubeControllerManagerArgs: map[string]*string{
						"--leader-elect-lease-duration": &[]string{"30s"}[0],
						"--leader-elect-renew-deadline": &[]string{"15s"}[0],
					},
					ExtraKubeSchedulerArgs: map[string]*string{
						"--leader-elect-lease-duration": &[]string{"30s"}[0],
						"--leader-elect-renew-deadline": &[]string{"15s"}[0],
					},
					Addons: []k8sinit.AddonConfiguration{
						{Name: "dns", Disable: false},
						{Name: "mayastor", Disable: false, Arguments: []string{"--default-pool-size", "20GB"}},
						{Name: "registry", Disable: true},
					},
					ContainerdRegistryConfigs: map[string]string{
						"docker.io": `server = "http://my.proxy:5000"`,
					},
					ExtraContainerdArgs: map[string]*string{
						"-l": &[]string{"debug"}[0],
					},
					ExtraContainerdEnv: map[string]*string{
						"http_proxy":  &[]string{"http://squid.internal:3128"}[0],
						"https_proxy": &[]string{"http://squid.internal:3128"}[0],
					},
					ExtraDqliteArgs: map[string]*string{
						"--disk-mode": &[]string{"true"}[0],
					},
					ExtraDqliteEnv: map[string]*string{
						"LIBRAFT_TRACE":   &[]string{"1"}[0],
						"LIBDQLITE_TRACE": &[]string{"1"}[0],
					},
					AddonRepositories: []k8sinit.AddonRepositoryConfiguration{
						{Name: "core", URL: "https://github.com/canonical/microk8s-core-addons"},
						{Name: "community", URL: "/snap/microk8s/current/addons/community", Reference: "1.26"},
					},
					ExtraConfigFiles: map[string]string{
						"flannel-network-mgr-config": `{"Network": "10.1.0.0/16", "Backend": {"Type": "vxlan"}}` + "\n",
					},
					PersistentClusterToken: "my-token",
					Join: k8sinit.JoinConfiguration{
						URL:    "10.0.0.10:25000/my-token/hash",
						Worker: true,
					},
				}},
			},
		},
		{
			name: "multi-part.yaml",
			expectConfiguration: k8sinit.MultiPartConfiguration{
				Parts: []*k8sinit.Configuration{
					{Version: "0.1.0", Addons: []k8sinit.AddonConfiguration{{Name: "dns"}}},
					{Version: "0.1.0", Addons: []k8sinit.AddonConfiguration{{Name: "rbac"}}},
				},
			},
		},
		{
			name: "multi-part-with-header.yaml",
			expectConfiguration: k8sinit.MultiPartConfiguration{
				Parts: []*k8sinit.Configuration{
					{Version: "0.1.0", Addons: []k8sinit.AddonConfiguration{{Name: "dns"}}},
					{Version: "0.1.0", Addons: []k8sinit.AddonConfiguration{{Name: "rbac"}}},
				},
			},
		},
		{
			name: "unknown-fields.yaml",
			expectConfiguration: k8sinit.MultiPartConfiguration{
				Parts: []*k8sinit.Configuration{{
					Version: "0.1.0",
				}},
			},
		},
		{name: "invalid-yaml.yaml", expectErr: true},
		{name: "invalid-schema.yaml", expectErr: true},
		{name: "version/newer.yaml", expectErr: true},
		{name: "version/non-semantic.yaml", expectErr: true},
		{name: "version/unsupported.yaml", expectErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			b, err := testdata.ReadFile(filepath.Join("testdata", "schema", filepath.Join(tc.name)))
			if err != nil {
				panic(err)
			}

			c, err := k8sinit.ParseMultiPartConfiguration(b)
			if tc.expectErr {
				g.Expect(err).NotTo(BeNil())
				g.Expect(c).To(BeZero())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(c).To(Equal(tc.expectConfiguration))
			}
		})
	}
}
