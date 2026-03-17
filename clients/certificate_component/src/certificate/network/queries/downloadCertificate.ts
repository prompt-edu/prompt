import { certificateAxiosInstance } from '../certificateServerConfig'

export const downloadOwnCertificate = async (coursePhaseId: string): Promise<Blob> => {
  const response = await certificateAxiosInstance.get(
    `certificate/api/course_phase/${coursePhaseId}/certificate/download`,
    { responseType: 'blob' },
  )
  return response.data
}

export const downloadStudentCertificate = async (
  coursePhaseId: string,
  studentId: string,
): Promise<Blob> => {
  const response = await certificateAxiosInstance.get(
    `certificate/api/course_phase/${coursePhaseId}/certificate/download/${studentId}`,
    { responseType: 'blob' },
  )
  return response.data
}

export const triggerBlobDownload = (blob: Blob, filename: string): void => {
  const url = window.URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  window.URL.revokeObjectURL(url)
}
