export interface StudyProgram {
  id: string
  name: string
  shortName: string | null
}

export interface CreateStudyProgram {
  name: string
  shortName: string
}

export interface UpdateStudyProgram {
  name: string
  shortName: string
}

export interface StudyProgramStudentCount {
  studyProgramID: string
  studentCount: number
}
