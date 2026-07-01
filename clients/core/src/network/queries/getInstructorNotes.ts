import { InstructorNote } from '@core/managementConsole/shared/interfaces/InstructorNote'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const getInstructorNotes = async (studentId: string): Promise<InstructorNote[]> => {
  try {
    return (
      await axiosInstance.get(`/api/instructor-notes/s/${studentId}`, {
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
