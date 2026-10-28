package provider

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
)

func TestIsBlockedAddr(t *testing.T) {
	blocked := []string{
		"127.0.0.1", "10.0.0.5", "172.16.0.1", "192.168.1.1", "169.254.169.254",
		"0.0.0.0", "100.64.0.1", "224.0.0.1", "255.255.255.255",
		"::", "::1", "fe80::1", "fd00:ec2::254", "ff02::1",
		"::ffff:10.0.0.1", "::ffff:169.254.169.254", "64:ff9b::a00:5", "2002:a00:5::1",
	}
	for _, raw := range blocked {
		if !isBlockedAddr(netip.MustParseAddr(raw)) {
			t.Errorf("%s is not blocked, but it reaches an internal or non-routable target", raw)
		}
	}

	public := []string{"8.8.8.8", "131.159.0.1", "2606:4700:4700::1111"}
	for _, raw := range public {
		if isBlockedAddr(netip.MustParseAddr(raw)) {
			t.Errorf("%s is blocked, but it is a public address", raw)
		}
	}
}

func TestUserFacingMessage(t *testing.T) {
	transport := &url.Error{Op: "Get", URL: "http://203.0.113.7:5432", Err: errors.New("dial tcp 203.0.113.7:5432: connect: connection refused")}
	if got := UserFacingMessage(transport); got != UnreachableMessage {
		t.Errorf("transport error leaks detail: %q", got)
	}
	if got := UserFacingMessage(errors.Join(ErrBlockedAddress, transport)); got != BlockedAddressMessage {
		t.Errorf("blocked address message = %q", got)
	}
	if got := UserFacingMessage(errors.New("gitlab answered 401")); got != "gitlab answered 401" {
		t.Errorf("non-transport error must pass through, got %q", got)
	}
}

func TestValidateURL(t *testing.T) {
	t.Setenv(EnvAllowedPrivateHosts, "")

	valid := []string{
		"https://gitlab.example.com",
		"https://gitlab.example.com/",
		"https://app.getoutline.com/api",
		" https://rancher.example.com:8443 ",
		"https://8.8.8.8",
	}
	for _, raw := range valid {
		if err := ValidateURL(raw); err != nil {
			t.Errorf("ValidateURL(%q) = %v, want nil", raw, err)
		}
	}

	invalid := []string{
		"http://gitlab.example.com",
		"ftp://gitlab.example.com",
		"gitlab.example.com",
		"https://",
		"https://user:secret@gitlab.example.com",
		"https://gitlab.example.com/?",
		"https://gitlab.example.com/#",
		"https://gitlab.example.com/api?x=1",
	}
	for _, raw := range invalid {
		if err := ValidateURL(raw); err == nil {
			t.Errorf("ValidateURL(%q) = nil, want an error", raw)
		}
	}

	internal := []string{
		"https://169.254.169.254/latest/meta-data",
		"https://10.0.0.5:5432",
		"https://[::1]/",
		"https://[fd00:ec2::254]/",
		"https://localhost",
		"https://api.localhost",
	}
	for _, raw := range internal {
		if err := ValidateURL(raw); !errors.Is(err, ErrBlockedAddress) {
			t.Errorf("ValidateURL(%q) = %v, want ErrBlockedAddress", raw, err)
		}
	}
}

func TestValidateURLHonorsAllowlist(t *testing.T) {
	t.Setenv(EnvAllowedPrivateHosts, "gitlab.internal, 10.0.0.0/8")

	for _, raw := range []string{"http://gitlab.internal", "https://GitLab.Internal.", "https://10.1.2.3", "http://10.1.2.3:8080"} {
		if err := ValidateURL(raw); err != nil {
			t.Errorf("ValidateURL(%q) = %v, want nil for an allowlisted host", raw, err)
		}
	}
	for _, raw := range []string{"https://192.168.1.1", "http://gitlab.example.com", "https://169.254.169.254"} {
		if err := ValidateURL(raw); err == nil {
			t.Errorf("ValidateURL(%q) = nil, want an error for a host outside the allowlist", raw)
		}
	}
}

func TestHTTPClientRefusesInternalAddress(t *testing.T) {
	t.Setenv(EnvAllowedPrivateHosts, "")

	var hits atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
	}))
	defer server.Close()

	// The test server listens on loopback, like an internal service would.
	_, err := NewHTTPClient().Get(server.URL)
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("GET %s = %v, want ErrBlockedAddress", server.URL, err)
	}
	if hits.Load() != 0 {
		t.Fatal("the internal server was reached")
	}

	// A hostname resolving to loopback is refused on the resolved address.
	localhostURL := strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
	if _, err := NewHTTPClient().Get(localhostURL); !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("GET %s = %v, want ErrBlockedAddress", localhostURL, err)
	}
}

func TestHTTPClientReachesAllowlistedHost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	for _, allowlist := range []string{"127.0.0.1", "127.0.0.0/8", "localhost"} {
		t.Run(allowlist, func(t *testing.T) {
			t.Setenv(EnvAllowedPrivateHosts, allowlist)
			target := server.URL
			if allowlist == "localhost" {
				target = strings.Replace(server.URL, "127.0.0.1", "localhost", 1)
			}
			resp, err := NewHTTPClient().Get(target)
			if err != nil {
				t.Fatalf("GET %s = %v, want success for an allowlisted host", target, err)
			}
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusNoContent {
				t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusNoContent)
			}
		})
	}
}

func TestHTTPClientRefusesRedirectToInternalAddress(t *testing.T) {
	// The redirecting server stands in for a public provider, so it is allowlisted; the
	// redirect target is not.
	t.Setenv(EnvAllowedPrivateHosts, "127.0.0.1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://169.254.169.254/latest/meta-data/", http.StatusFound)
	}))
	defer server.Close()

	_, err := NewHTTPClient().Get(server.URL)
	if !errors.Is(err, ErrBlockedAddress) {
		t.Fatalf("GET with a redirect to the metadata endpoint = %v, want ErrBlockedAddress", err)
	}
}

func TestHTTPClientRefusesRedirectToPlainHTTP(t *testing.T) {
	t.Setenv(EnvAllowedPrivateHosts, "127.0.0.1")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://gitlab.example.com/api/v4/user", http.StatusFound)
	}))
	defer server.Close()

	if _, err := NewHTTPClient().Get(server.URL); err == nil || !strings.Contains(err.Error(), "must use https") {
		t.Fatalf("GET with a redirect to plain http = %v, want a https error", err)
	}
}

func TestHTTPClientStopsAfterTooManyRedirects(t *testing.T) {
	t.Setenv(EnvAllowedPrivateHosts, "127.0.0.1")

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, server.URL+"/again", http.StatusFound)
	}))
	defer server.Close()

	if _, err := NewHTTPClient().Get(server.URL); err == nil || !strings.Contains(err.Error(), "redirects") {
		t.Fatalf("GET with a redirect loop = %v, want a redirect limit error", err)
	}
}
