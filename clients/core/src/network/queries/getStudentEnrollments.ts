import { axiosInstance } from '@/network/configService'
import { StudentEnrollments } from '@core/managementConsole/shared/interfaces/StudentEnrollment'

export const getStudentEnrollments = async (studentId: string): Promise<StudentEnrollments> => {
  try {
    return (
      await axiosInstance.get(`/api/students/${studentId}/enrollments`, {
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
