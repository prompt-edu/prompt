import { createAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'

// EXAMPLE_HOST is injected via core's env.js but not yet part of the
// @tumaet/prompt-shared-state EnvType, so it is read from window.env directly.
const exampleServer = (window.env as { EXAMPLE_HOST?: string }).EXAMPLE_HOST ?? ''

const exampleAxiosInstance = createAuthenticatedAxiosInstance(exampleServer)

export { exampleAxiosInstance }
