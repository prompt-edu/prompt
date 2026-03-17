import { axiosInstance } from '@/network/configService'

interface FileDownloadUrlResponse {
  downloadUrl: string
}

export const getApplicationFileDownloadUrl = async (
  coursePhaseId: string,
  fileId: string,
): Promise<string> => {
  const response = await axiosInstance.get<FileDownloadUrlResponse>(
    `/api/applications/${coursePhaseId}/files/${fileId}/download-url`,
  )
  return response.data.downloadUrl
}
