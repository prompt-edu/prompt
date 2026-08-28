package provider

import (
	"fmt"

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
