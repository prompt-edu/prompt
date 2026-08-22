import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'

const certificateAxiosInstance = createAuthenticatedAxiosInstance(env.CERTIFICATE_HOST)

export { certificateAxiosInstance }
