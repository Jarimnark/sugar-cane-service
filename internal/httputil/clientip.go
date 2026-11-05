package httputil

import (
	"net"
	"net/http"
)

// ClientIP returns the client IP from r.RemoteAddr with the port stripped.
// If RemoteAddr is unparseable, it is returned verbatim as a last resort.
func ClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
