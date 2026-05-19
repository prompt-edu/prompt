import { axiosInstance } from '@tumaet/prompt-shared-state'
import { CoursePhaseGraphItem } from '../../managementConsole/courseConfigurator/interfaces/coursePhaseGraphItem'

export const getCoursePhaseGraph = async (courseID: string): Promise<CoursePhaseGraphItem[]> => {
  try {
    return (await axiosInstance.get(`/api/courses/${courseID}/phase_graph`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
