package provider

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"unicode"

	log "github.com/sirupsen/logrus"
)

// StatusError describes a failed upstream request. Callers react to the status and code
// rather than matching on the message, which is prose and changes.
type StatusError struct {
	Provider string
	Method   string
	Path     string
	Status   int
	// Code is the machine-readable error code the response body carried, when it had
	// one. Rancher answers 422 for a duplicate and for a validation error alike, and the
	// code is the only thing that separates them.
	Code string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("%s %s %s: HTTP %d", e.Provider, e.Method, e.Path, e.Status)
}

// StatusOf reports the upstream HTTP status carried by err, if it carries one.
func StatusOf(err error) (int, bool) {
	var statusErr *StatusError
	if errors.As(err, &statusErr) {
		return statusErr.Status, true
	}
	return 0, false
}

// HasStatus reports whether err came from an upstream response with the given status.
func HasStatus(err error, status int) bool {
	got, ok := StatusOf(err)
	return ok && got == status
}

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

	return &StatusError{
		Provider: providerType,
		Method:   method,
		Path:     path,
		Status:   status,
		Code:     upstreamCode(body),
	}
}

// upstreamCode reads a machine-readable error code out of a JSON error body. Only a
// top-level string code is taken; no free-form upstream text reaches the caller, since
// this error is persisted on the instance and rendered in the UI.
func upstreamCode(body []byte) string {
	var decoded struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ""
	}
	return decoded.Code
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
