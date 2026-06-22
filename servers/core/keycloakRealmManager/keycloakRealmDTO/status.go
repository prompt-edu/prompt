package keycloakRealmDTO

// KeycloakStatus reports whether the configured Keycloak service account is
// usable for the operations the course-team management feature depends on.
// Each capability is an independent sub-check; Healthy is true only when all
// of them pass.
type KeycloakStatus struct {
	ServiceName  string          `json:"serviceName"`
	Healthy      bool            `json:"healthy"`
	Capabilities map[string]bool `json:"capabilities"`
}

// Keycloak capability keys reported by the status endpoint. Kept stable -
// the frontend maps these to display labels.
const (
	// CapabilityKeycloakLogin reports whether the service account can obtain
	// an access token via client credentials.
	CapabilityKeycloakLogin = "keycloak.login"
	// CapabilityKeycloakReadUsers reports whether the service account can list
	// realm users (requires Keycloak's view-users role).
	CapabilityKeycloakReadUsers = "keycloak.read_users"
	// CapabilityKeycloakReadGroups reports whether the service account can list
	// realm groups (requires Keycloak's view-groups role).
	CapabilityKeycloakReadGroups = "keycloak.read_groups"
)
