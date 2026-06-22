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

// Capability keys. Kept stable - the frontend maps these to display labels.
const (
	CapabilityKeycloakLogin      = "keycloak.login"
	CapabilityKeycloakReadUsers  = "keycloak.read_users"
	CapabilityKeycloakReadGroups = "keycloak.read_groups"
)
