package utiltest

import "net"

// MockCIDR is a mock net.Addr
type MockCIDR struct {
	CIDR string
}

// Network implements net.Addr
func (v *MockCIDR) Network() string {
	return ""
}

// String implements net.Addr
func (v *MockCIDR) String() string {
	return v.CIDR
}

var _ net.Addr = &MockCIDR{}
