import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'

const interviewAxiosInstance = createAuthenticatedAxiosInstance(env.INTERVIEW_HOST)

export { interviewAxiosInstance }
