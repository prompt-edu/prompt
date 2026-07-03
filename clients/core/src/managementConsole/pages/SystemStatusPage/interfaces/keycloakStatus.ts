export interface KeycloakStatus {
  healthy: boolean
  capabilities: Record<string, boolean>
}

// Keys mirror the constants in servers/core/keycloakRealmManager/keycloakRealmDTO/status.go
export const KEYCLOAK_CAPABILITY_LABELS: Record<string, string> = {
  'keycloak.login': 'Client-credentials login',
  'keycloak.read_users': 'Read users (view-users)',
  'keycloak.read_groups': 'Read groups (view-groups)',
}
