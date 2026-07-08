import type {
  CreateNoteTag,
  NoteTag,
} from '@core/managementConsole/shared/interfaces/InstructorNote'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const postNoteTag = async (tag: CreateNoteTag): Promise<NoteTag> => {
  try {
    return (
      await axiosInstance.post('/api/instructor-notes/tags', tag, {
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
