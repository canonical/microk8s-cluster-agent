package util_test

import (
	"net"
	"testing"

	"github.com/canonical/microk8s-cluster-agent/pkg/util"
)

func TestGetRemoteHost(t *testing.T) {
	for _, tc := range []struct {
		hostname      string
		resolvesTo    []net.IP
		remoteAddress string
		expected      string
	}{
		{
			hostname:      "host-with-dns",
			resolvesTo:    []net.IP{{1, 1, 1, 1}},
			remoteAddress: "1.1.1.1:31412",
			expected:      "host-with-dns",
		},
		{
			hostname:      "host-with-multiple-dns",
			resolvesTo:    []net.IP{{1, 1, 1, 1}, {1, 1, 1, 2}},
			remoteAddress: "1.1.1.1:31412",
			expected:      "host-with-multiple-dns",
		},
		{
			hostname:      "host-without-dns",
			resolvesTo:    nil,
			remoteAddress: "1.1.1.2:31241",
			expected:      "1.1.1.2",
		},
		{
			hostname:      "host-with-wrong-dns",
			resolvesTo:    []net.IP{{1, 1, 1, 3}},
			remoteAddress: "1.1.1.4:41232",
			expected:      "1.1.1.4",
		},
	} {
		t.Run(tc.hostname, func(t *testing.T) {
			lookupIP := func(string) ([]net.IP, error) { return tc.resolvesTo, nil }

			if host := util.GetRemoteHost(lookupIP, tc.hostname, tc.remoteAddress); host != tc.expected {
				t.Fatalf("Expected remote host to be %q but it was %q instead", tc.expected, host)
			}
		})
	}
}
