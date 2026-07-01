import type {
  NoteTag,
  UpdateNoteTag,
} from '@core/managementConsole/shared/interfaces/InstructorNote'
import { axiosInstance } from '@tumaet/prompt-shared-state'

export const putNoteTag = async (tagId: string, tag: UpdateNoteTag): Promise<NoteTag> => {
  try {
    return (
      await axiosInstance.put(`/api/instructor-notes/tags/${tagId}`, tag, {
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
