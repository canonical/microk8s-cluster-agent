package snap_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap"
	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
)

func TestAddons(t *testing.T) {
	t.Run("EnableDisable", func(t *testing.T) {
		runner := &utiltest.MockRunner{}
		s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(runner.Run))

		s.EnableAddon(context.Background(), "dns")
		s.EnableAddon(context.Background(), "dns", "10.0.0.2")
		s.DisableAddon(context.Background(), "rbac")
		s.DisableAddon(context.Background(), "storage", "--destroy")

		expectedCommands := []string{
			"testdata/microk8s-enable.wrapper dns",
			"testdata/microk8s-enable.wrapper dns 10.0.0.2",
			"testdata/microk8s-disable.wrapper rbac",
			"testdata/microk8s-disable.wrapper storage --destroy",
		}
		if !reflect.DeepEqual(expectedCommands, runner.CalledWithCommand) {
			t.Fatalf("Expected commands %#v, but received %#v", expectedCommands, runner.CalledWithCommand)
		}
	})

	t.Run("AddRepository", func(t *testing.T) {
		runner := &utiltest.MockRunner{}
		s := snap.NewSnap("testdata", "testdata", snap.WithCommandRunner(runner.Run))

		s.AddAddonsRepository(context.Background(), "core", "/snap/microk8s/current/addons/core", "", false)
		s.AddAddonsRepository(context.Background(), "core", "/snap/microk8s/current/addons/core", "", true)
		s.AddAddonsRepository(context.Background(), "core", "https://github.com/canonical/microk8s-core-addons", "1.26", false)
		s.AddAddonsRepository(context.Background(), "community", "https://github.com/canonical/microk8s-community-addons", "1.26", true)

		expectedCommands := []string{
			"testdata/microk8s-addons.wrapper repo add core /snap/microk8s/current/addons/core",
			"testdata/microk8s-addons.wrapper repo add core /snap/microk8s/current/addons/core --force",
			"testdata/microk8s-addons.wrapper repo add core https://github.com/canonical/microk8s-core-addons --reference 1.26",
			"testdata/microk8s-addons.wrapper repo add community https://github.com/canonical/microk8s-community-addons --reference 1.26 --force",
		}
		if !reflect.DeepEqual(expectedCommands, runner.CalledWithCommand) {
			t.Fatalf("Expected commands %#v, but received %#v", expectedCommands, runner.CalledWithCommand)
		}
	})
}
