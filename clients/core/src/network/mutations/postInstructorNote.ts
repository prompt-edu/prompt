import { axiosInstance } from '@/network/configService'
import {
  CreateInstructorNote,
  InstructorNote,
} from '@core/managementConsole/shared/interfaces/InstructorNote'

export const postInstructorNote = async (
  studentId: string,
  note: CreateInstructorNote,
): Promise<InstructorNote[]> => {
  try {
    return (
      await axiosInstance.post(`/api/instructor-notes/s/${studentId}`, note, {
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
