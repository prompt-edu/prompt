import { CertificateStatus } from '../../interfaces/participant'
import { certificateAxiosInstance } from '../certificateServerConfig'

export const getCertificateStatus = async (coursePhaseId: string): Promise<CertificateStatus> => {
  const response = await certificateAxiosInstance.get(
    `certificate/api/course_phase/${coursePhaseId}/certificate/status`,
  )
  return response.data
}
