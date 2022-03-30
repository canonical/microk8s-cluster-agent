package snaputil_test

import (
	"context"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

func TestMaybePatchCalicoAutoDetectionMethod(t *testing.T) {
	snap := &mock.Snap{
		CNIYaml: `some contents here and there. value: "first-found"`,
	}

	if err := snaputil.MaybePatchCalicoAutoDetectionMethod(context.Background(), snap, "10.10.10.10", true); err != nil {
		t.Fatalf("Expected no errors but received %q", err)
	}

	expectedYaml := `some contents here and there. value: "can-reach=10.10.10.10"`
	if snap.CNIYaml != expectedYaml {
		t.Fatalf("expected CNI yaml to be %q but it was %q", expectedYaml, snap.CNIYaml)
	}
	if len(snap.ApplyCNICalled) != 1 {
		t.Fatalf("expected ApplyCNI to be called but it was not")
	}
}
