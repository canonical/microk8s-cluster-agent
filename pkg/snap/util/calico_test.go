package snaputil_test

import (
	"context"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/snap/mock"
	snaputil "github.com/canonical/microk8s-cluster-agent/pkg/snap/util"
	. "github.com/onsi/gomega"
)

func TestMaybePatchCalicoAutoDetectionMethod(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		oldYAML              string
		canReachHost         string
		expectYAML           string
		expectApplyCNICalled int
	}{
		{
			name: "IPv4 Address",
			oldYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "10.10.10.10",
			expectYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=10.10.10.10"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			expectApplyCNICalled: 1,
		},
		{
			name: "IPv6 Address",
			oldYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
			expectYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
			expectApplyCNICalled: 1,
		},
		{
			name: "IPv4 Address not first-found",
			oldYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=1.1.1.1"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "10.10.10.10",
			expectYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "can-reach=1.1.1.1"
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		},
		{
			name: "IPv6 Address not first-found",
			oldYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
			expectYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"
- name: IP6_AUTODETECTION_METHOD
  value: "can-reach=2001:0db8:85a3:0000:0000:8a2e:0370:7334"`,
		},
		{
			name: "IPv4 Address no autodetection",
			oldYAML: `
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "1.1.1.1",
			expectYAML: `
- name: IP6_AUTODETECTION_METHOD
  value: "first-found"`,
		},
		{
			name: "IPv6 Address no autodetection",
			oldYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"`,
			canReachHost: "2001:0db8:85a3:0000:0000:8a2e:0370:7335",
			expectYAML: `
- name: IP_AUTODETECTION_METHOD
  value: "first-found"`,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := NewWithT(t)
			snap := &mock.Snap{
				CNIYaml: tc.oldYAML,
			}

			err := snaputil.MaybePatchCalicoAutoDetectionMethod(context.Background(), snap, tc.canReachHost, true)
			g.Expect(err).To(BeNil())
			g.Expect(snap.CNIYaml).To(Equal(tc.expectYAML))
			g.Expect(snap.ApplyCNICalled).To(HaveLen(tc.expectApplyCNICalled))
		})
	}
}
