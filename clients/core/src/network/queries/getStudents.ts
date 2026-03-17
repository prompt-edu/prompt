import { Student } from '@tumaet/prompt-shared-state'
import { axiosInstance } from '@/network/configService'

export const getStudents = async (): Promise<Student[]> => {
  try {
    return (
      await axiosInstance.get('/api/students/', {
        headers: {
          'Content-Type': 'application/json',
        },
      })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
