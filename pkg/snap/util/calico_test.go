package snaputil_test

import (
	"context"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
)

func TestMaybePatchCalicoAutoDetectionMethod(t *testing.T) {
	tests := []struct {
		name         string
		cniYaml      string
		canReachHost string
		expectedYaml string
	}{
		{
			name: "IPv4 Address",
			cniYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "10.10.10.10",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=10.10.10.10"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		},
		{
			name: "IPv6 Address",
			cniYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
		},
		{
			name: "IPv4 Address not first-found",
			cniYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=1.1.1.1"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "10.10.10.10",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=1.1.1.1"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		},
		{
			name: "IPv6 Address not first-found",
			cniYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
		},
		{
			name: "IPv4 Address no autodetection",
			cniYaml: `
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "1.1.1.1",
			expectedYaml: `
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		},
		{
			name: "IPv6 Address no autodetection",
			cniYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snap := &mock.Snap{
				CNIYaml: test.cniYaml,
			}

			if err := snaputil.MaybePatchCalicoAutoDetectionMethod(context.Background(), snap, test.canReachHost, true); err != nil {
				t.Fatalf("Expected no errors but received %q", err)
			}

			if snap.CNIYaml != test.expectedYaml {
				t.Fatalf("expected CNI yaml to be %q but it was %q", test.expectedYaml, snap.CNIYaml)
			}

			if len(snap.ApplyCNICalled) != 1 {
				t.Fatalf("expected ApplyCNI to be called but it was not")
			}
		})
	}
}
