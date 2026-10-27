package keycloak

import (
	"os"
	"testing"

	"github.com/prompt-edu/prompt/servers/infrastructure_setup/provider"
)

// TestMain trusts loopback for the test servers these tests start. The provider HTTP
// client refuses internal addresses unless they are allowlisted, as in production.
func TestMain(m *testing.M) {
	if err := os.Setenv(provider.EnvAllowedPrivateHosts, "127.0.0.1,::1"); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
