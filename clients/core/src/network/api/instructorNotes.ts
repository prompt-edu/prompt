import type {
  CreateInstructorNote,
  CreateNoteTag,
  InstructorNote,
  NoteTag,
  UpdateNoteTag,
} from '@core/managementConsole/shared/interfaces/InstructorNote'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/instructor-notes`
const tagsPath = `${path}/tags`

export const instructorNotes = {
  ofStudent: (studentID: string): Promise<InstructorNote[]> =>
    coreRequest.get(`${path}/s/${studentID}`),

  create: (studentID: string, note: CreateInstructorNote): Promise<InstructorNote[]> =>
    coreRequest.post(`${path}/s/${studentID}`, note),

  remove: (noteID: string): Promise<InstructorNote> => coreRequest.del(`${path}/${noteID}`),

  listTags: (): Promise<NoteTag[]> => coreRequest.get(tagsPath),

  createTag: (tag: CreateNoteTag): Promise<NoteTag> => coreRequest.post(tagsPath, tag),

  updateTag: (tagID: string, tag: UpdateNoteTag): Promise<NoteTag> =>
    coreRequest.put(`${tagsPath}/${tagID}`, tag),

  removeTag: (tagID: string): Promise<void> => coreRequest.del(`${tagsPath}/${tagID}`),
}
