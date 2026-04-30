import { axiosInstance } from '@tumaet/prompt-shared-state'
import { NoteTag } from '@core/managementConsole/shared/interfaces/InstructorNote'

export const getNoteTags = async (): Promise<NoteTag[]> => {
  try {
    return (
      await axiosInstance.get('/api/instructor-notes/tags', {
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
