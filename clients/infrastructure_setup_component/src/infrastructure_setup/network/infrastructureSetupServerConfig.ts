import { createAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'

// INFRASTRUCTURE_SETUP_HOST is injected by core's env.js. It is read off window.env
// rather than through the shared `env` object because that object normalizes away every
// key it does not know, and the @tumaet/prompt-shared-state EnvType does not carry this
// host yet - the presentation remote reads its own host the same way.
const infrastructureSetupHost =
  typeof window === 'undefined'
    ? ''
    : ((window.env as { INFRASTRUCTURE_SETUP_HOST?: string }).INFRASTRUCTURE_SETUP_HOST ?? '')

// The shared helper resolves the base URL inside the request interceptor and injects the
// JWT, so a host that only becomes available after this module is evaluated still applies.
const infrastructureSetupAxiosInstance = createAuthenticatedAxiosInstance(infrastructureSetupHost)

export { infrastructureSetupAxiosInstance }
