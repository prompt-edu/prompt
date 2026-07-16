// Centralized environment config. Defaults target the e2e stack's host-published
// ports (4000 / 18081 / 18090, chosen to avoid clashing with a dev stack) so
// running Playwright directly on the host (cd e2e && npx playwright test) works
// against a running stack with no env vars; the containerized runner overrides
// these via docker-compose.e2e.yml.

export const BASE_URL = process.env.BASE_URL ?? 'http://localhost:4000'
export const KEYCLOAK_URL = process.env.KEYCLOAK_URL ?? 'http://localhost:18081'
export const CORE_API_URL = process.env.CORE_API_URL ?? 'http://localhost:18090'
export const KEYCLOAK_REALM = process.env.KEYCLOAK_REALM_NAME ?? 'prompt'

// Public Keycloak client used by the browser app (direct access grants enabled).
export const KEYCLOAK_CLIENT_ID = 'prompt-client'

// Phase-module API prefix on the browser origin (BASE_URL); proxied to the
// phase server by e2e/nginx/client-core.conf, mirroring prod Traefik routing.
export const SELF_TEAM_ALLOCATION_API = '/self-team-allocation/api'
export const ASSESSMENT_API = '/assessment/api'
export const INTERVIEW_API = '/interview/api'
export const CERTIFICATE_API = '/certificate/api'
export const PRESENTATION_API = '/presentation/api'

export const tokenEndpoint = () =>
  `${KEYCLOAK_URL}/realms/${KEYCLOAK_REALM}/protocol/openid-connect/token`
