import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'

const teamAllocationAxiosInstance = createAuthenticatedAxiosInstance(env.TEAM_ALLOCATION_HOST)

export { teamAllocationAxiosInstance }
