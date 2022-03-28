package util

import "net"

// GetRemoteHost returns the hostname that should be used for communicating with the joining node.
// The endpoint is either the hostname (if it can be resolved), or the remote IP address, as read from the HTTP request.
// lookupIP is net.LookupIP.
// If the hostname resolves to a different IP than remoteAddress, then the IP from remoteAddress is returned.
func GetRemoteHost(lookupIP func(string) ([]net.IP, error), hostname string, remoteAddress string) string {
	remoteHost, _, _ := net.SplitHostPort(remoteAddress)
	if ips, err := lookupIP(hostname); err == nil && len(ips) > 0 {
		remoteHostIP := net.ParseIP(remoteHost)
		for _, ip := range ips {
			if ip.Equal(remoteHostIP) {
				return hostname
			}
		}
	}
	return remoteHost
}
