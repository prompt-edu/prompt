import type { InstructorNote } from '@core/managementConsole/shared/interfaces/InstructorNote'
import { axiosInstance } from '@tumaet/prompt-shared-state'

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
