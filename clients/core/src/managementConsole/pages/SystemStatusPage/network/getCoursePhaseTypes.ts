import { axiosInstance } from '@/network/configService'
import { CoursePhaseType } from '../interfaces/coursePhaseType'

export async function getCoursePhaseTypes(): Promise<CoursePhaseType[]> {
  const res = await axiosInstance.get<CoursePhaseType[]>('/api/course_phase_types')
  return res.data
}
