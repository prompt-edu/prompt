import { axiosInstance } from '@/network/configService'
import { NoteTag, UpdateNoteTag } from '@core/managementConsole/shared/interfaces/InstructorNote'

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
