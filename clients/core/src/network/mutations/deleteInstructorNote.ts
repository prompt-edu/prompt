import { axiosInstance } from '@/network/configService'
import { InstructorNote } from '@core/managementConsole/shared/interfaces/InstructorNote'

export const deleteInstructorNote = async (noteId: string): Promise<InstructorNote> => {
  try {
    return (
      await axiosInstance.delete(`/api/instructor-notes/${noteId}`, {
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
