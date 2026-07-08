import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { CoursePhaseType } from '../interfaces/coursePhaseType'

export async function getCoursePhaseTypes(forSelf?: boolean): Promise<CoursePhaseType[]> {
  const res = await axiosInstance.get<CoursePhaseType[]>('/api/course_phase_types', {
    params: forSelf ? { for_self: true } : undefined,
  })
  return res.data
}
