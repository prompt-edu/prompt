package provider

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"net/url"
	"os"
	"strings"
	"syscall"
)

// EnvAllowedPrivateHosts lists the provider hosts an operator explicitly trusts although
// they resolve to a private, loopback or otherwise internal address, for example a
// GitLab instance that only exists on the university network. Entries are comma
// separated and are hostnames, IP addresses or CIDR ranges. It is empty by default, so
// every internal address is refused.
const EnvAllowedPrivateHosts = "INFRASTRUCTURE_SETUP_ALLOWED_PRIVATE_HOSTS"

// ErrBlockedAddress marks a provider request refused because its target is an internal
// address. Provider URLs are set by lecturers, so without this guard a lecturer could
// point the service at the cloud metadata endpoint or scan the internal network from it.
var ErrBlockedAddress = errors.New("provider address is not allowed")

// BlockedAddressMessage is the text shown to a lecturer for an ErrBlockedAddress. It
// names neither the resolved address nor the port, so it reveals nothing about the
// internal network.
const BlockedAddressMessage = "the provider URL points to an internal address, which is not allowed"

// UnreachableMessage is the text shown to a lecturer for a transport failure. The raw
// error names the dialled address and tells a refused connection from a timeout or an
// unknown host, which would reveal which hosts and ports answer.
const UnreachableMessage = "could not reach the provider, check the URL"

// UserFacingMessage turns a provider error into text that is safe to show a lecturer.
// Upstream status and configuration errors carry no network detail and pass through.
func UserFacingMessage(err error) string {
	if errors.Is(err, ErrBlockedAddress) {
		return BlockedAddressMessage
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return UnreachableMessage
	}
	return err.Error()
}

// blockedPrefixes are the ranges that are not covered by netip's classification but
// still reach internal or non-routable targets.
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),      // "this network"
	netip.MustParsePrefix("100.64.0.0/10"),  // carrier-grade NAT
	netip.MustParsePrefix("192.0.0.0/24"),   // IETF protocol assignments
	netip.MustParsePrefix("198.18.0.0/15"),  // benchmarking
	netip.MustParsePrefix("240.0.0.0/4"),    // reserved
	netip.MustParsePrefix("::/96"),          // IPv4-compatible IPv6
	netip.MustParsePrefix("64:ff9b::/96"),   // NAT64, embeds an IPv4 target
	netip.MustParsePrefix("64:ff9b:1::/48"), // local-use NAT64
	netip.MustParsePrefix("2002::/16"),      // 6to4, embeds an IPv4 target
	netip.MustParsePrefix("2001:db8::/32"),  // documentation
	netip.MustParsePrefix("100::/64"),       // discard-only
	netip.MustParsePrefix("2001::/32"),      // Teredo, embeds an IPv4 target
	netip.MustParsePrefix("fec0::/10"),      // deprecated site-local
}

// isBlockedAddr reports whether addr is an internal or non-routable address: loopback,
// private (RFC 1918 and fc00::/7, which includes the fd00:ec2::254 metadata address),
// link-local (which includes the 169.254.169.254 metadata address), unspecified,
// multicast, or one of blockedPrefixes.
func isBlockedAddr(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsPrivate() {
		return true
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// addressPolicy decides which provider targets may be dialled.
type addressPolicy struct {
	hosts    map[string]struct{}
	prefixes []netip.Prefix
}

// loadAddressPolicy reads EnvAllowedPrivateHosts. It is read for every client, so a
// changed environment applies without a restart of the provider registry.
func loadAddressPolicy() addressPolicy {
	return parseAddressPolicy(os.Getenv(EnvAllowedPrivateHosts))
}

func parseAddressPolicy(raw string) addressPolicy {
	policy := addressPolicy{hosts: map[string]struct{}{}}
	for _, entry := range strings.Split(raw, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			policy.prefixes = append(policy.prefixes, prefix.Masked())
			continue
		}
		if addr, err := netip.ParseAddr(entry); err == nil {
			addr = addr.Unmap()
			policy.prefixes = append(policy.prefixes, netip.PrefixFrom(addr, addr.BitLen()))
			continue
		}
		policy.hosts[normalizeHost(entry)] = struct{}{}
	}
	return policy
}

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

// hostAllowed reports whether host is an explicitly trusted hostname.
func (p addressPolicy) hostAllowed(host string) bool {
	_, ok := p.hosts[normalizeHost(host)]
	return ok
}

// addrAllowed reports whether addr may be dialled.
func (p addressPolicy) addrAllowed(addr netip.Addr) bool {
	addr = addr.Unmap()
	if !isBlockedAddr(addr) {
		return true
	}
	for _, prefix := range p.prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// trusted reports whether host is on the allowlist, either by name or as an allowed
// literal address. Only trusted hosts may be reached over plain http.
func (p addressPolicy) trusted(host string) bool {
	if p.hostAllowed(host) {
		return true
	}
	addr, err := netip.ParseAddr(host)
	if err != nil {
		return false
	}
	addr = addr.Unmap()
	for _, prefix := range p.prefixes {
		if prefix.Contains(addr) {
			return true
		}
	}
	return false
}

// ValidateURL checks a lecturer-supplied provider base URL before it is stored.
//
// The URL must be absolute https with a host and must not carry credentials, a query or
// a fragment: providers append API paths to the raw string, so a "?" or "#" would turn
// those paths into an ignored suffix. A literal internal address or a localhost name is
// rejected here already. Hostnames are not resolved, because the dial guard checks the
// address a request actually connects to, which also covers DNS changes after saving.
// Plain http is accepted only for a host on the EnvAllowedPrivateHosts allowlist.
func ValidateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if strings.ContainsAny(raw, "?#") {
		return errors.New("must not contain a query or fragment")
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return errors.New("is not a valid URL")
	}
	return loadAddressPolicy().validateTarget(parsed)
}

// validateTarget checks the scheme, host and credentials of a URL a provider request
// goes to. It runs on stored base URLs and on every redirect target.
func (p addressPolicy) validateTarget(target *url.URL) error {
	host := target.Hostname()
	if target.Opaque != "" || host == "" {
		return errors.New("must be an absolute URL with a host")
	}
	switch target.Scheme {
	case "https":
	case "http":
		if !p.trusted(host) {
			return errors.New("must use https")
		}
	default:
		return errors.New("must use https")
	}
	if target.User != nil {
		return errors.New("must not contain credentials")
	}
	if p.trusted(host) {
		return nil
	}
	normalized := normalizeHost(host)
	if normalized == "localhost" || strings.HasSuffix(normalized, ".localhost") {
		return ErrBlockedAddress
	}
	if addr, err := netip.ParseAddr(host); err == nil && !p.addrAllowed(addr) {
		return ErrBlockedAddress
	}
	return nil
}

// dialContext returns a dial function that refuses internal addresses.
//
// The check runs in the dialer's Control hook, on the address the socket is about to
// connect to after DNS resolution. That covers hostnames resolving to internal
// addresses, DNS rebinding between validation and use, and every redirect hop, since
// each new connection passes through here. A hostname on the allowlist skips the check.
func (p addressPolicy) dialContext() func(ctx context.Context, network, address string) (net.Conn, error) {
	guarded := &net.Dialer{
		Timeout:   dialTimeout,
		KeepAlive: dialKeepAlive,
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return ErrBlockedAddress
			}
			addr, err := netip.ParseAddr(host)
			if err != nil || !p.addrAllowed(addr) {
				return ErrBlockedAddress
			}
			return nil
		},
	}
	trusted := &net.Dialer{Timeout: dialTimeout, KeepAlive: dialKeepAlive}

	return func(ctx context.Context, network, address string) (net.Conn, error) {
		if host, _, err := net.SplitHostPort(address); err == nil && p.hostAllowed(host) {
			return trusted.DialContext(ctx, network, address)
		}
		return guarded.DialContext(ctx, network, address)
	}
}
