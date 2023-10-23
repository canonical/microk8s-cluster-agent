package snap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

var (
	dqliteCrt         = "DQLITE CERTIFICATE DATA"
	dqliteKey         = "DQLITE KEY DATA"
	dqliteInfoYaml    = "DQLITE INFO YAML"
	dqliteClusterYaml = "DQLITE CLUSTER YAML"
	caCrt             = "CA CERTIFICATE DATA"
	caKey             = "CA KEY DATA"
	saKey             = "SERVICE ACCOUNT KEY DATA"
	cniYaml           = "CNI DATA"
)

func TestFiles(t *testing.T) {
	// Create test data
	for file, contents := range map[string]string{
		"testdata/var/kubernetes/backend/cluster.crt":  dqliteCrt,
		"testdata/var/kubernetes/backend/cluster.key":  dqliteKey,
		"testdata/var/kubernetes/backend/info.yaml":    dqliteInfoYaml,
		"testdata/var/kubernetes/backend/cluster.yaml": dqliteClusterYaml,
		"testdata/certs/ca.crt":                        caCrt,
		"testdata/certs/ca.key":                        caKey,
		"testdata/certs/serviceaccount.key":            saKey,
		"testdata/args/cni-network/cni.yaml":           cniYaml,
	} {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %s", err)
		}
		if err := os.WriteFile(file, []byte(contents), 0660); err != nil {
			t.Fatalf("Failed to create test file: %s", err)
		}
		defer os.RemoveAll(filepath.Dir(file))
	}

	s := snap.NewSnap("testdata", "testdata", "testdata")

	for _, tc := range []struct {
		name     string
		retrieve func() (string, error)
		expected string
	}{
		{name: "CA", retrieve: s.ReadCA, expected: caCrt},
		{name: "CAKey", retrieve: s.ReadCAKey, expected: caKey},
		{name: "ServiceAccountKey", retrieve: s.ReadServiceAccountKey, expected: saKey},
		{name: "DqliteInfoYaml", retrieve: s.ReadDqliteInfoYaml, expected: dqliteInfoYaml},
		{name: "DqliteClusterYaml", retrieve: s.ReadDqliteClusterYaml, expected: dqliteClusterYaml},
		{name: "DqliteCrt", retrieve: s.ReadDqliteCert, expected: dqliteCrt},
		{name: "DqliteKey", retrieve: s.ReadDqliteKey, expected: dqliteKey},
		{name: "CNI", retrieve: s.ReadCNIYaml, expected: cniYaml},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, err := tc.retrieve()
			if err != nil {
				t.Fatalf("expected error to be nil, but it was %q instead", err)
			}
			if v != tc.expected {
				t.Fatalf("expected %s to be %q, but it was %q instead", tc.name, tc.expected, v)
			}
		})
	}

	for _, tc := range []struct {
		name  string
		write func([]byte) error
		file  string
	}{
		{name: "CNI", write: s.WriteCNIYaml, file: "testdata/args/cni-network/cni.yaml"},
		{name: "DqliteUpdateYaml", write: s.WriteDqliteUpdateYaml, file: "testdata/var/kubernetes/backend/update.yaml"},
		{name: "WriteCSRConfig", write: s.WriteCSRConfig, file: "testdata/certs/csr.conf.template"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			magicVal := util.NewRandomString(util.Alpha, 10)
			if err := tc.write([]byte(magicVal)); err != nil {
				t.Fatalf("expected error to be nil, but it was %q instead", err)
			}

			v, err := util.ReadFile(tc.file)
			if err != nil {
				t.Fatalf("expected error to be nil, but it was %q instead", err)
			}
			if v != magicVal {
				t.Fatalf("expected contents of %q to be %q, but they were %q instead", tc.file, magicVal, v)
			}
		})
	}
}
