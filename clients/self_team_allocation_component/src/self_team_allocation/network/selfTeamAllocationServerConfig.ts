import { createAuthenticatedAxiosInstance, env } from '@tumaet/prompt-shared-state'

const selfTeamAllocationAxiosInstance = createAuthenticatedAxiosInstance(
  env.SELF_TEAM_ALLOCATION_HOST,
)

export { selfTeamAllocationAxiosInstance }
