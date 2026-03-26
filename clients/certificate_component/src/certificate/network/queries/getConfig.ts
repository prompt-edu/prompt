import { CoursePhaseConfig } from '../../interfaces/participant'
import { certificateAxiosInstance } from '../certificateServerConfig'

export const getConfig = async (coursePhaseId: string): Promise<CoursePhaseConfig> => {
  const response = await certificateAxiosInstance.get(
    `certificate/api/course_phase/${coursePhaseId}/config`,
  )
  return response.data
}

export const updateConfig = async (
  coursePhaseId: string,
  templateContent: string,
): Promise<CoursePhaseConfig> => {
  const response = await certificateAxiosInstance.put(
    `certificate/api/course_phase/${coursePhaseId}/config`,
    { templateContent },
  )
  return response.data
}

export const updateReleaseDate = async (
  coursePhaseId: string,
  releaseDate: string | null,
): Promise<CoursePhaseConfig> => {
  const response = await certificateAxiosInstance.put(
    `certificate/api/course_phase/${coursePhaseId}/config/release-date`,
    { releaseDate },
  )
  return response.data
}
