import { axiosInstance, type Student } from '@tumaet/prompt-shared-state'

export const getStudentsOfCoursePhase = async (coursePhaseID: string): Promise<Student[]> => {
  return (await axiosInstance.get(`/api/course_phases/${coursePhaseID}/participations/students`))
    .data
}
