package snap_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
)

func TestRunUpgrade(t *testing.T) {

	for file, contents := range map[string]string{
		"testdata/credentials/callback-token.txt":                      "valid-token",
		"testdata/upgrade-scripts/001-custom-upgrade/prepare-node.sh":  "",
		"testdata/upgrade-scripts/001-custom-upgrade/commit-node.sh":   "",
		"testdata/upgrade-scripts/001-custom-upgrade/rollback-node.sh": "",
	} {
		if err := os.MkdirAll(filepath.Dir(file), 0755); err != nil {
			t.Fatalf("Failed to create test directory: %s", err)
		}
		if err := os.WriteFile(file, []byte(contents), 0660); err != nil {
			t.Fatalf("Failed to create test file: %s", err)
		}
	}
	defer os.RemoveAll("testdata/credentials")
	defer os.RemoveAll("testdata/upgrade-scripts")

	runner := &utiltest.MockRunner{}
	s := snap.NewSnap("testdata", "testdata", runner.Run)

	t.Run("Invalid", func(t *testing.T) {
		for _, tc := range []struct {
			upgrade string
			phase   string
		}{
			{upgrade: "001-custom-upgrade", phase: "invalid-phase"},
			{upgrade: "999-invalid-upgrade", phase: "prepare"},
		} {
			t.Run(fmt.Sprintf("%s/%s", tc.upgrade, tc.phase), func(t *testing.T) {
				err := s.RunUpgrade(context.Background(), tc.upgrade, tc.phase)
				if err == nil {
					t.Fatal("Expected an error but did not receive any")
				}
				if len(runner.CalledWithCommand) > 0 {
					t.Fatalf("Expected no commands to be called, but received %#v", runner.CalledWithCommand)
				}
			})
		}
	})

	t.Run("Success", func(t *testing.T) {
		for _, phase := range []string{"prepare", "commit", "rollback"} {
			t.Run(phase, func(t *testing.T) {

				runner := &utiltest.MockRunner{}
				s := snap.NewSnap("testdata", "testdata", runner.Run)

				err := s.RunUpgrade(context.Background(), "001-custom-upgrade", phase)
				if err != nil {
					t.Fatalf("Expected no errors but received %q", err)
				}
				expectedCommands := []string{fmt.Sprintf("testdata/upgrade-scripts/001-custom-upgrade/%s-node.sh", phase)}
				if !reflect.DeepEqual(expectedCommands, runner.CalledWithCommand) {
					t.Fatalf("Expected commands %#v, but received %#v", expectedCommands, runner.CalledWithCommand)
				}
			})
		}
	})

}
