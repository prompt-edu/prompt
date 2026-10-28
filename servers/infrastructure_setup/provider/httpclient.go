package provider

import (
	"errors"
	"fmt"
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

const (
	dialTimeout   = 10 * time.Second
	dialKeepAlive = 30 * time.Second
	// maxRedirects bounds how many redirects one provider request follows.
	maxRedirects = 5
)

// NewHTTPClient returns the HTTP client providers use for upstream calls.
//
// Provider URLs come from lecturers, so the client refuses to connect to internal
// addresses (see addressPolicy.dialContext) and re-validates every redirect target. It
// ignores HTTP(S)_PROXY: a proxy would make the dial guard check the proxy instead of
// the provider.
func NewHTTPClient() *http.Client {
	policy := loadAddressPolicy()

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.DialContext = policy.dialContext()

	return &http.Client{
		Timeout:   RequestTimeout,
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			if err := policy.validateTarget(req.URL); err != nil {
				if errors.Is(err, ErrBlockedAddress) {
					return err
				}
				return fmt.Errorf("redirect target %w", err)
			}
			return nil
		},
	}
}
