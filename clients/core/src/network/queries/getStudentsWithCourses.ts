import { axiosInstance } from '@/network/configService'

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
  studyProgram: string
  courses: StudentCourseParticipation[]
  noteTags: StudentNoteTag[]
}

export const getStudentsWithCourses = async (): Promise<StudentWithCourses[]> => {
  try {
    return (
      await axiosInstance.get('/api/students/with-courses', {
        headers: {
          'Content-Type': 'application/json-path+json',
        },
      })
    ).data
  } catch (err) {
    console.error(err)
    throw err
  }
}
