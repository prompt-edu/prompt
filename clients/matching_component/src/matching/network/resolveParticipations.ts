import type {
  CoursePhaseParticipationWithStudent,
  DataResolution,
} from '@tumaet/prompt-shared-state'
import axios from 'axios'

/**
 * Client-side equivalent of the prompt-sdk `ResolveAllParticipations` + merge.
 *
 * Matching has no backend, so it must resolve the REST `resolutions` core hands
 * out itself: for each resolution it fetches the provider endpoint and merges the
 * value into every participation's `prevData`, keyed by `courseParticipationID`.
 * The user's Keycloak token is forwarded, exactly like the server-side SDK does.
 */

interface ResolvedDto {
  dtoName: string
  valuesByParticipation: Record<string, unknown>
}

const buildResolutionURL = (resolution: DataResolution): string => {
  const base = resolution.baseURL.replace(/\/+$/, '')
  const endpointPath = resolution.endpointPath.replace(/^\/+|\/+$/g, '')
  return `${base}/course_phase/${resolution.coursePhaseID}/${endpointPath}`
}

const authHeaders = (): Record<string, string> => {
  const token = typeof localStorage !== 'undefined' ? localStorage.getItem('jwt_token') : null
  return token ? { Authorization: `Bearer ${token}` } : {}
}

/**
 * Pure merge step: fold the resolved provider values into `prevData`. Extracted so
 * it can be unit-tested without any network access.
 */
export const mergeResolutions = (
  participations: CoursePhaseParticipationWithStudent[],
  resolvedDtos: ResolvedDto[],
): CoursePhaseParticipationWithStudent[] => {
  if (resolvedDtos.length === 0) {
    return participations
  }

  return participations.map((participation) => {
    const prevData = { ...(participation.prevData ?? {}) }
    for (const { dtoName, valuesByParticipation } of resolvedDtos) {
      const value = valuesByParticipation[participation.courseParticipationID]
      if (value !== undefined) {
        prevData[dtoName] = value
      }
    }
    return { ...participation, prevData }
  })
}

/**
 * Fetch every resolution and merge the results into the participations. Resolutions
 * that fail to load are skipped so a single unavailable provider does not blank out
 * the whole page.
 */
export const resolveParticipations = async (
  participations: CoursePhaseParticipationWithStudent[],
  resolutions: DataResolution[],
): Promise<CoursePhaseParticipationWithStudent[]> => {
  if (!resolutions || resolutions.length === 0) {
    return participations
  }

  const resolvedDtos = await Promise.all(
    resolutions.map(async (resolution): Promise<ResolvedDto> => {
      const valuesByParticipation: Record<string, unknown> = {}
      try {
        const response = await axios.get<Array<Record<string, unknown>>>(
          buildResolutionURL(resolution),
          { headers: authHeaders() },
        )
        for (const item of response.data ?? []) {
          const id = item.courseParticipationID
          if (typeof id === 'string') {
            valuesByParticipation[id] = item[resolution.dtoName]
          }
        }
      } catch (err) {
        console.error(`Failed to resolve "${resolution.dtoName}" from ${resolution.baseURL}`, err)
      }
      return { dtoName: resolution.dtoName, valuesByParticipation }
    }),
  )

  return mergeResolutions(participations, resolvedDtos)
}
