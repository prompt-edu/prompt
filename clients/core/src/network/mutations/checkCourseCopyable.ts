import { axiosInstance } from '@tumaet/prompt-shared-state'
import type { CheckCourseCopyableResponse } from '../../managementConsole/courseOverview/interfaces/checkCourseCopyableResponse'

export const checkCourseCopyable = async (
  courseID: string,
): Promise<CheckCourseCopyableResponse> => {
  try {
    const response = await axiosInstance.get(`/api/courses/${courseID}/copyable`)
    return response.data
  } catch (err) {
    console.error(err)
    throw err
  }
}
