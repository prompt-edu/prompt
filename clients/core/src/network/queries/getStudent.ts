import { axiosInstance } from '@/network/configService'
import { Student } from '@tumaet/prompt-shared-state'

export const getStudent = async (studentId: string): Promise<Student> => {
  try {
    return (
      await axiosInstance.get(`/api/students/${studentId}`, {
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
