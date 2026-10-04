import type { CoursePhaseType } from '@core/managementConsole/courseConfigurator/interfaces/coursePhaseType'
import type { CoursePhaseType as ServiceCoursePhaseType } from '@core/managementConsole/pages/SystemStatusPage/interfaces/coursePhaseType'
import type {
  CoursePhaseWithMetaData,
  CreateCoursePhase,
  UpdateCoursePhase,
} from '@tumaet/prompt-shared-state'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/course_phases`
const typesPath = `${API_PREFIX}/course_phase_types`

export const coursePhases = {
  byID: (coursePhaseID: string): Promise<CoursePhaseWithMetaData> =>
    coreRequest.get(`${path}/${coursePhaseID}`),

  create: async (coursePhase: CreateCoursePhase): Promise<string | undefined> =>
    (await coreRequest.post<{ id?: string }>(`${path}/course/${coursePhase.courseID}`, coursePhase))
      .id,

  update: async (coursePhase: UpdateCoursePhase): Promise<string | undefined> =>
    (await coreRequest.put<{ id?: string }>(`${path}/${coursePhase.id}`, coursePhase)).id,

  remove: (coursePhaseID: string): Promise<void> => coreRequest.del(`${path}/${coursePhaseID}`),

  listTypes: (): Promise<CoursePhaseType[]> => coreRequest.get(typesPath),

  // The same route as `listTypes`, read as the richer shape the system status page needs, and
  // optionally narrowed to the phase types the caller may configure
  listTypesForScope: (forSelf?: boolean): Promise<ServiceCoursePhaseType[]> =>
    coreRequest.get(typesPath, forSelf ? { params: { for_self: true } } : undefined),
}
