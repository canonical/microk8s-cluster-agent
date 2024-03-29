package snap_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
)

func TestLock(t *testing.T) {
	s := snap.NewSnap("testdata", "testdata", "testdata")
	if err := os.MkdirAll("testdata/var/lock", 0755); err != nil {
		t.Fatalf("Failed to create directory: %s", err)
	}
	defer os.RemoveAll("testdata/var")
	for _, tc := range []struct {
		name    string
		file    string
		hasLock func() bool
	}{
		{name: "kubelite", file: "lite.lock", hasLock: s.HasKubeliteLock},
		{name: "dqlite", file: "ha-cluster", hasLock: s.HasDqliteLock},
		{name: "cert-reissue", file: "no-cert-reissue", hasLock: s.HasNoCertsReissueLock},
	} {
		t.Run(tc.name, func(t *testing.T) {
			lockFile := filepath.Join("testdata", "var", "lock", tc.file)
			if err := os.Remove(lockFile); err != nil && !os.IsNotExist(err) {
				t.Fatalf("Failed to remove %s: %s", lockFile, err)
			}
			if tc.hasLock() {
				t.Fatal("Expected not to have lock but we do")
			}
			if _, err := os.Create(lockFile); err != nil {
				t.Fatalf("Failed to create %s: %s", lockFile, err)
			}
			if !tc.hasLock() {
				t.Fatal("Expected to have lock but we do not")
			}
			os.Remove(lockFile)
		})
	}
}
