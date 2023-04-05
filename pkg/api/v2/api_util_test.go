package v2

import (
	"fmt"
	"net"
	"testing"

	utiltest "github.com/canonical/microk8s-cluster-agent/pkg/util/test"
	. "github.com/onsi/gomega"
)

func TestStdlibPreconditions(t *testing.T) {
	g := NewWithT(t)
	addrs, err := net.InterfaceAddrs()
	g.Expect(err).To(BeNil(), "net.InterfaceAddrs() must not fail")

	for _, addr := range addrs {
		ip, subnet, err := net.ParseCIDR(addr.String())
		g.Expect(err).To(BeNil(), "net.ParseCIDR() must not fail for %v", addr.String())
		g.Expect(ip).ToNot(BeNil(), "Address for %v must not be nil", addr.String())
		g.Expect(subnet).ToNot(BeNil(), "Subnet for %v must not be nil", addr.String())
	}
}

func TestFindMatchingBindAddress(t *testing.T) {
	t.Run("FallbackOnFailure", func(t *testing.T) {
		g := NewWithT(t)

		var interfaceAddrsCalled bool
		a := API{
			InterfaceAddrs: func() ([]net.Addr, error) {
				interfaceAddrsCalled = true
				return nil, fmt.Errorf("some error")
			},
		}

		addr, err := a.findMatchingBindAddress("10.0.0.10:25000")
		g.Expect(interfaceAddrsCalled).To(BeTrue())
		g.Expect(err).To(BeNil())
		g.Expect(addr).To(Equal("10.0.0.10"))
	})

	t.Run("FailOnMissing", func(t *testing.T) {
		g := NewWithT(t)
		a := API{InterfaceAddrs: net.InterfaceAddrs}

		// 1.1.1.1 should almost never be a host address, this test will fail otherwise
		addr, err := a.findMatchingBindAddress("1.1.1.1:25000")
		g.Expect(err).ToNot(BeNil())
		g.Expect(addr).To(BeEmpty())
	})

	t.Run("HandleVirtualIP", func(t *testing.T) {
		a := API{
			InterfaceAddrs: func() ([]net.Addr, error) {
				return []net.Addr{
					&utiltest.MockCIDR{CIDR: "127.0.0.1/8"},
					&utiltest.MockCIDR{CIDR: "10.0.0.10/16"},
					&utiltest.MockCIDR{CIDR: "10.10.0.10/16"},
					&utiltest.MockCIDR{CIDR: "10.0.100.100/32"},
					&utiltest.MockCIDR{CIDR: "192.168.100.100/32"},
				}, nil
			},
		}
		t.Run("UseInterfaceIP", func(t *testing.T) {
			g := NewWithT(t)

			addr, err := a.findMatchingBindAddress("10.0.0.10:25000")
			g.Expect(err).To(BeNil())
			g.Expect(addr).To(Equal("10.0.0.10"))
		})

		t.Run("UseVirtualIP", func(t *testing.T) {
			g := NewWithT(t)

			addr, err := a.findMatchingBindAddress("10.0.100.100:25000")
			g.Expect(err).To(BeNil())
			g.Expect(addr).To(Equal("10.0.0.10"))
		})

		t.Run("FallbackToVirtualIPIfSubnetNotFound", func(t *testing.T) {
			g := NewWithT(t)

			addr, err := a.findMatchingBindAddress("192.168.100.100:25000")
			g.Expect(err).To(BeNil())
			g.Expect(addr).To(Equal("192.168.100.100"))
		})
	})
}
