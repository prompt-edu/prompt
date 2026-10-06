import type {
  CreateStudyProgram,
  StudyProgram,
  StudyProgramStudentCount,
  UpdateStudyProgram,
} from '@core/managementConsole/shared/interfaces/StudyProgram'
import { API_PREFIX, coreRequest, publicRequest } from '../client'

const path = `${API_PREFIX}/study-programs`

export const studyPrograms = {
  list: (): Promise<StudyProgram[]> => publicRequest.get(path),

  studentCounts: (): Promise<StudyProgramStudentCount[]> =>
    coreRequest.get(`${path}/student-counts`),

  create: (studyProgram: CreateStudyProgram): Promise<StudyProgram> =>
    coreRequest.post(path, studyProgram),

  update: (studyProgramID: string, studyProgram: UpdateStudyProgram): Promise<StudyProgram> =>
    coreRequest.put(`${path}/${studyProgramID}`, studyProgram),

  remove: (studyProgramID: string): Promise<void> => coreRequest.del(`${path}/${studyProgramID}`),
}
