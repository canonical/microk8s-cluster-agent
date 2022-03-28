package v1

import "net"

// API implements the v1 API.
type API struct {
	// LookupIP is net.LookupIP.
	LookupIP func(string) ([]net.IP, error)
}
