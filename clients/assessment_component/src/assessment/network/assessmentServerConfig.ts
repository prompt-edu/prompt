import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'

const assessmentAxiosInstance = createAuthenticatedAxiosInstance(env.ASSESSMENT_HOST)

export { assessmentAxiosInstance }
