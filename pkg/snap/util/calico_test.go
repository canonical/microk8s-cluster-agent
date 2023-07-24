package snaputil_test

import (
	"context"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

func TestMaybePatchCalicoAutoDetectionMethod(t *testing.T) {
	snap := &mock.Snap{
		CNIYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
	}

	if err := snaputil.MaybePatchCalicoAutoDetectionMethod(context.Background(), snap, "10.10.10.10", true); err != nil {
		t.Fatalf("Expected no errors but received %q", err)
	}

	expectedYaml := `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=10.10.10.10"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`
	if snap.CNIYaml != expectedYaml {
		t.Fatalf("expected CNI yaml to be %q but it was %q", expectedYaml, snap.CNIYaml)
	}

	if err := snaputil.MaybePatchCalicoAutoDetectionMethod(context.Background(), snap, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true); err != nil {
		t.Fatalf("Expected no errors but received %q", err)
	}
	expectedYaml = `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=10.10.10.10"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`
	if snap.CNIYaml != expectedYaml {
		t.Fatalf("expected CNI yaml to be %q but it was %q", expectedYaml, snap.CNIYaml)
	}

	if len(snap.ApplyCNICalled) != 2 {
		t.Fatalf("expected ApplyCNI to be called but it was not")
	}
}
