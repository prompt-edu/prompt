// Centralized environment config. Defaults target the e2e stack's host-published
// ports (4000 / 18081 / 18090 — chosen to avoid clashing with a dev stack) so
// host mode (`make test-e2e-local`) works with no env vars; the containerized
// runner overrides these via docker-compose.e2e.yml.

export const BASE_URL = process.env.BASE_URL ?? 'http://localhost:4000'
export const KEYCLOAK_URL = process.env.KEYCLOAK_URL ?? 'http://localhost:18081'
export const CORE_API_URL = process.env.CORE_API_URL ?? 'http://localhost:18090'
export const KEYCLOAK_REALM = process.env.KEYCLOAK_REALM_NAME ?? 'prompt'

// Public Keycloak client used by the browser app (direct access grants enabled).
export const KEYCLOAK_CLIENT_ID = 'prompt-client'

export const tokenEndpoint = () =>
  `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token`
