import { axiosInstance, type Student } from '@tumaet/prompt-shared-state'

export const searchStudents = async (searchString: string): Promise<Student[]> => {
  return (await axiosInstance.get(`/api/students/search/${encodeURIComponent(searchString)}`)).data
}
