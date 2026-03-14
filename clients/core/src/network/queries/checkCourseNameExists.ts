import { axiosInstance } from '@/network/configService'

export const checkCourseNameExists = async (
  name: string,
  semesterTag: string,
): Promise<boolean> => {
  const response = await axiosInstance.get('/api/courses/check-name', {
    params: { name, semesterTag },
  })
  return response.data.exists
}
