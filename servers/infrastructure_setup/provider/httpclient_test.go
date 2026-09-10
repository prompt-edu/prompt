package provider

import "testing"

func TestNewHTTPClientHasARequestTimeout(t *testing.T) {
	if timeout := NewHTTPClient().Timeout; timeout == 0 {
		t.Fatal("client has no timeout, so an unresponsive provider host would park a worker until the run expires")
	}
}
