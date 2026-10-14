package feature

import (
	"errors"
	"fmt"
	"strings"
)

type Policy struct {
	ContentRetentionDays int
	// EU AI Act Annex III: restricted instead of deleted on erasure, kept at least 183 days.
	HighRisk bool
}

// Every call names its feature, so there is no call without a retention.
var registry = map[string]Policy{
	"assessment.*": {ContentRetentionDays: 730, HighRisk: true},
}

var ErrNotAllowed = errors.New("feature not allowed")

func Lookup(name string) (Policy, bool) {
	if policy, ok := registry[name]; ok {
		return policy, true
	}
	prefix, rest, found := strings.Cut(name, ".")
	if !found || prefix == "" || rest == "" {
		return Policy{}, false
	}
	policy, ok := registry[prefix+".*"]
	return policy, ok
}

// Strictest covers stored features the registry no longer knows.
func Strictest() Policy {
	return Policy{ContentRetentionDays: MaxContentRetentionDays(), HighRisk: true}
}

func PolicyOf(name string) Policy {
	if policy, ok := Lookup(name); ok {
		return policy
	}
	return Strictest()
}

func MaxContentRetentionDays() int {
	longest := 0
	for _, policy := range registry {
		longest = max(longest, policy.ContentRetentionDays)
	}
	return longest
}

func Slug(phaseTypeName string) string {
	return strings.ToLower(strings.Join(strings.Fields(phaseTypeName), "-"))
}

func CheckFor(name, phaseTypeName string) error {
	if _, ok := Lookup(name); !ok {
		return fmt.Errorf("%w: %q is not registered", ErrNotAllowed, name)
	}
	if prefix, _, _ := strings.Cut(name, "."); prefix != Slug(phaseTypeName) {
		return fmt.Errorf("%w: %q does not belong to phase type %q", ErrNotAllowed, name, phaseTypeName)
	}
	return nil
}
