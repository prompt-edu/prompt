import { axiosInstance, type Student } from '@tumaet/prompt-shared-state'

export const getStudentsOfCoursePhase = async (coursePhaseID: string): Promise<Student[]> => {
  try {
    const response = await axiosInstance.get(
      `/api/course_phases/${coursePhaseID}/participations/students`,
    )
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
