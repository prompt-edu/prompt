import { axiosInstance } from '@/network/configService'

export const deleteApplicationFile = async (
  coursePhaseId: string,
  fileId: string,
): Promise<void> => {
  if (!coursePhaseId) {
    throw new Error('coursePhaseId is required to delete a file')
  }
  if (!fileId) {
    throw new Error('fileId is required to delete a file')
  }

  await axiosInstance.delete(`/api/apply/authenticated/${coursePhaseId}/files/${fileId}`)
}
