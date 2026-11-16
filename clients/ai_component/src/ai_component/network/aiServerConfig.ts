import { createAuthenticatedAxiosInstance } from '@tumaet/prompt-shared-state'

// AI_HOST is injected via core's env.js but not part of the @tumaet/prompt-shared-state EnvType,
// so it is read from window.env directly.
const aiServer = (window.env as { AI_HOST?: string }).AI_HOST ?? ''

export const aiAxiosInstance = createAuthenticatedAxiosInstance(aiServer)

export const aiPhasePath = (coursePhaseId: string) => `/ai/api/course_phase/${coursePhaseId}`
