import { ParticipantWithDownloadStatus } from '../../interfaces/participant'
import { certificateAxiosInstance } from '../certificateServerConfig'

export const getParticipants = async (
  coursePhaseId: string,
): Promise<ParticipantWithDownloadStatus[]> => {
  const response = await certificateAxiosInstance.get(
    `certificate/api/course_phase/${coursePhaseId}/participants`,
  )
  return response.data
}
