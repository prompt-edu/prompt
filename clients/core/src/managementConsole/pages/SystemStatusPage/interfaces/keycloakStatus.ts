export interface KeycloakStatus {
  serviceName: string
  healthy: boolean
  capabilities: Record<string, boolean>
}

// Mirrors the constants in servers/core/keycloakRealmManager/keycloakRealmDTO/status.go
export const CAPABILITY_KEYCLOAK_LOGIN = 'keycloak.login'
export const CAPABILITY_KEYCLOAK_READ_USERS = 'keycloak.read_users'
export const CAPABILITY_KEYCLOAK_READ_GROUPS = 'keycloak.read_groups'

export const KEYCLOAK_CAPABILITY_LABELS: Record<string, string> = {
  [CAPABILITY_KEYCLOAK_LOGIN]: 'Client-credentials login',
  [CAPABILITY_KEYCLOAK_READ_USERS]: 'Read users (view-users)',
  [CAPABILITY_KEYCLOAK_READ_GROUPS]: 'Read groups (view-groups)',
}
