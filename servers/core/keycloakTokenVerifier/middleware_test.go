package keycloakTokenVerifier

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
)

func TestLogTokenVerificationFailure(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantLevel log.Level
	}{
		{"expired token", &oidc.TokenExpiredError{Expiry: time.Now()}, log.DebugLevel},
		{"wrapped expired token", fmt.Errorf("verify: %w", &oidc.TokenExpiredError{Expiry: time.Now()}), log.DebugLevel},
		{"generic failure", errors.New("token signature is invalid"), log.ErrorLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prevLevel := log.GetLevel()
			prevHooks := log.StandardLogger().Hooks
			hook := test.NewGlobal()
			log.SetLevel(log.DebugLevel)
			t.Cleanup(func() {
				log.SetLevel(prevLevel)
				log.StandardLogger().ReplaceHooks(prevHooks)
			})

			logTokenVerificationFailure(tt.err)

			entry := hook.LastEntry()
			if entry == nil {
				t.Fatal("expected a log entry, got none")
			} else if entry.Level != tt.wantLevel {
				t.Fatalf("expected level %v, got %v", tt.wantLevel, entry.Level)
			}
		})
	}
}
