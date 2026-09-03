package provider

import (
	"fmt"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

// maxLoggedBodyBytes bounds how much of an upstream response body reaches the log.
const maxLoggedBodyBytes = 512

// HTTPError describes a failed upstream request by method, path and status, without the
// response body.
//
// The returned message is persisted on the resource instance and rendered in the phase
// UI, and upstream bodies can carry tokens, internal hostnames or other users' data. The
// body is logged at debug level instead, truncated.
func HTTPError(providerType, method, path string, status int, body []byte) error {
	logged := body
	if len(logged) > maxLoggedBodyBytes {
		logged = logged[:maxLoggedBodyBytes]
	}
	log.WithFields(log.Fields{
		"provider": providerType,
		"method":   method,
		"path":     path,
		"status":   status,
		"body":     string(logged),
	}).Debug("upstream request failed")

	return fmt.Errorf("%s %s %s: HTTP %d", providerType, method, path, status)
}

// maxUpstreamReasonRunes bounds a reason quoted from an upstream payload.
const maxUpstreamReasonRunes = 200

// UpstreamReason sanitises a per-item failure reason taken from a response payload so it
// can be shown to a lecturer.
//
// Some endpoints report a per-member outcome only in the body, and that reason is worth
// surfacing. It is still upstream-controlled text landing in a persisted, UI-rendered
// field, so it is stripped of control characters, collapsed onto one line and truncated.
func UpstreamReason(reason string) string {
	cleaned := strings.Map(func(r rune) rune {
		if r == '\n' || r == '\r' || r == '\t' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, reason)
	cleaned = strings.Join(strings.Fields(cleaned), " ")

	// Truncation is by rune, not by byte: a cut inside a multi-byte rune leaves invalid
	// UTF-8, and this string is stored in a text column, which Postgres would reject -
	// leaving the instance unmarked and stuck in_progress.
	runes := []rune(cleaned)
	if len(runes) > maxUpstreamReasonRunes {
		return string(runes[:maxUpstreamReasonRunes]) + "..."
	}
	return cleaned
}
