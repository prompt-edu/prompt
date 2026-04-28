import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CoursePhaseType } from '../../managementConsole/courseConfigurator/interfaces/coursePhaseType'

export const getAllCoursePhaseTypes = async (): Promise<CoursePhaseType[]> => {
  try {
    return (await axiosInstance.get(`/api/course_phase_types`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
