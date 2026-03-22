import { axiosInstance } from '@/network/configService'

export const requestStudentDataExport = async (): Promise<PrivacyExport> => {
  try {
    return (
      await axiosInstance.post('/api/privacy/data-export', {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export enum ExportStatus {
  pending = 'pending',
  complete = 'complete',
  failed = 'failed',
}

export interface PrivacyExport {
  id: string
  userID: string
  studentID: string
  status: ExportStatus
  date_created: string
  valid_until: string
  documents: PrivacyExportDocument[]
}
export interface PrivacyExportDocument {
  id: string
  date_created: string
  source_name: string
  object_key: string
  status: ExportStatus
  file_size: number | null
}

export const getStudentDataExportStatus = async (exportID: string): Promise<PrivacyExport> => {
  try {
    return (await axiosInstance.get(`/api/privacy/data-export/${exportID}`)).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export const getExportDocDownloadURL = async (exportID: string, docID: string): Promise<string> => {
  try {
    const response = await axiosInstance.get(
      `/api/privacy/data-export/${exportID}/docs/${docID}/download-url`,
    )
    return response.data.downloadUrl
  } catch (err) {
    console.error(err)
    throw err
  }
}

export type LatestExportResponse =
  | { status: 'exists'; export: PrivacyExport }
  | { status: 'rate_limited'; retry_after: string }
  | { status: 'ready' }

export const getLatestStudentDataExport = async (): Promise<LatestExportResponse> => {
  try {
    const response = await axiosInstance.get('/api/privacy/data-export')
    if (response.status === 204) return { status: 'ready' }
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
