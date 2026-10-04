package provider

import (
	"net/http"
	"time"
)

// RequestTimeout bounds a single upstream request.
//
// Without it a provider host that accepts the connection and never answers (a firewall
// dropping the response, a hung TLS handshake) parks a worker goroutine until the whole
// run's context expires, so one unreachable provider consumes the entire run and the
// retry logic is never reached.
const RequestTimeout = 30 * time.Second

// NewHTTPClient returns the HTTP client providers use for upstream calls.
func NewHTTPClient() *http.Client {
	return &http.Client{Timeout: RequestTimeout}
}
