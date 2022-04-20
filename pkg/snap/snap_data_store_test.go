package snap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

func TestDataStore(t *testing.T) {
	s := snap.NewSnap("testdata", "testdata", nil)
	if err := os.MkdirAll("testdata/var/lock", 0755); err != nil {
		t.Fatalf("Failed to create directory: %s", err)
	}
	defer os.RemoveAll("testdata/var")
	for _, tc := range []struct {
		file  string
		store snap.DataStore
	}{
		{file: "ha-cluster", store: snap.DqliteDataStore},
		{file: "-", store: snap.SingleNodeEtcdDataStore},
		{file: "ha-etcd.lock", store: snap.EtcdDataStore},
	} {
		t.Run(string(tc.store), func(t *testing.T) {
			lockFile := filepath.Join("testdata", "var", "lock", tc.file)
			if err := os.Remove(lockFile); err != nil && !os.IsNotExist(err) {
				t.Fatalf("Failed to remove %s: %s", lockFile, err)
			}
			if _, err := os.Create(lockFile); err != nil {
				t.Fatalf("Failed to create %s: %s", lockFile, err)
			}
			if s.GetDataStore() != tc.store {
				t.Fatalf("Expected the store to be %q but it is %q instead", tc.store, s.GetDataStore())
			}
			os.Remove(lockFile)
		})
	}
}
