package auth

import (
	"conf"
	"net"
)

// Tells ir remote address allowed to access particular URL
func CanAccess(url conf.Url, address string) bool {
	host, _, _ := net.SplitHostPort(address)
	ip := net.ParseIP(host)
	for _, n := range url.Networks {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
