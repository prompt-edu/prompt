import axios from 'axios'
import { certificateAxiosInstance } from '../certificateServerConfig'

export interface PreviewError {
  error: string
  compilerOutput?: string
}

export const previewCertificate = async (coursePhaseId: string): Promise<Blob> => {
  try {
    const response = await certificateAxiosInstance.get(
      `certificate/api/course_phase/${coursePhaseId}/certificate/preview`,
      { responseType: 'blob' },
    )
    return response.data
  } catch (err: unknown) {
    // When responseType is 'blob', error responses are also blobs.
    // We need to read them as text and parse the JSON.
    if (axios.isAxiosError(err) && err.response?.data instanceof Blob) {
      const text = await err.response.data.text()
      try {
        const parsed: PreviewError = JSON.parse(text)
        throw parsed
      } catch (parseErr) {
        if ((parseErr as PreviewError).compilerOutput !== undefined) {
          throw parseErr
        }
      }
    }
    throw err
  }
}
