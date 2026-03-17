import { axiosInstance } from '@/network/configService'
import { CreateNoteTag, NoteTag } from '@core/managementConsole/shared/interfaces/InstructorNote'

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
