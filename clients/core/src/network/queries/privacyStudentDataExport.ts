import { axiosInstance } from '@/network/configService'

export const requestStudentDataExport = async (): Promise<PrivacyExport> => {
  try {
    return (await axiosInstance.post('/api/privacy/data-export')).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

export enum ExportStatus {
  pending = 'pending',
  complete = 'complete',
  no_data = 'no_data',
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
  status: ExportStatus
  file_size: number | null
  downloaded_at: string | null
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

export interface AdminExportDoc {
  source_name: string
  status: ExportStatus
  downloaded: boolean
}

export interface AdminPrivacyExport {
  id: string
  status: ExportStatus
  date_created: string
  valid_until: string
  docs: AdminExportDoc[]
}

export const getAllExports = async (): Promise<AdminPrivacyExport[]> => {
  try {
    return (await axiosInstance.get('/api/privacy/admin/data-exports')).data
  } catch (err) {
    console.error(err)
    throw err
  }
}

// DEV ONLY - delete later
export const devResetExports = async (): Promise<void> => {
  await axiosInstance.delete('/api/privacy/dev/reset-exports')
}

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
