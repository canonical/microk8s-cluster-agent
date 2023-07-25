package snaputil_test

import (
	"context"
	"fmt"
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
  value: value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectedYaml: `
- name: IP_AUTODETECTION_METHOD
  value: value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
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

func TestReplaceAfter(t *testing.T) {
	tests := []struct {
		input       string
		after       string
		target      string
		replacement string
		expected    string
		errExpected bool
	}{
		// Test cases
		{"ipv4 -value=first-found IPV6 -value first-found first-found", "IPV6", "first-found", "can-after=a3a3:a3a3:a3r3:3af3", "ipv4 -value=first-found IPV6 -value can-after=a3a3:a3a3:a3r3:3af3 first-found", false},
		{"ipv4 -value=first-found IPV6 -value first-found first-found", "IPV4", "first-found", "can-after=a3a3:a3a3:a3r3:3af3", "", true}, // After string not found
		{"ipv4 -value=first-found IPV6 -value first-found first-found", "IPV6", "not-found", "can-after=a3a3:a3a3:a3r3:3af3", "", true},   // Target string not found after "after"
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("ReplaceAfter(%q, %q, %q, %q)", test.input, test.after, test.target, test.replacement), func(t *testing.T) {
			actual, err := snaputil.ReplaceAfter(test.input, test.after, test.target, test.replacement)

			if test.errExpected {
				if err == nil {
					t.Errorf("Expected an error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if actual != test.expected {
				t.Errorf("Expected %q, but got %q", test.expected, actual)
			}
		})
	}
}
