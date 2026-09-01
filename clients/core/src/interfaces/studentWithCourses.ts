import type { Gender, StudyDegree } from '@tumaet/prompt-shared-state'

export interface StudentCourseParticipation {
  courseId: string
  courseName: string
  studentReadableData: object
}

export interface StudentNoteTag {
  id: string
  name: string
  color: string
}

export interface StudentWithCourses {
  id: string
  firstName: string
  lastName: string
  email: string
  hasUniversityAccount: boolean
  currentSemester?: number
  gender: Gender
  nationality: string
  studyDegree: StudyDegree
  studyProgram: string
  lastModified: string
  courses: StudentCourseParticipation[]
  noteTags: StudentNoteTag[]
}
