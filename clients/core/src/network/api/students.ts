import type { StudentWithCourses } from '@core/interfaces/studentWithCourses'
import type { StudentEnrollments } from '@core/managementConsole/shared/interfaces/StudentEnrollment'
import type { Student } from '@tumaet/prompt-shared-state'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/students`

export const students = {
  byID: (studentID: string): Promise<Student> => coreRequest.get(`${path}/${studentID}`),

  enrollments: (studentID: string): Promise<StudentEnrollments> =>
    coreRequest.get(`${path}/${studentID}/enrollments`),

  withCourses: (): Promise<StudentWithCourses[]> => coreRequest.get(`${path}/with-courses`),

  search: (searchString: string): Promise<Student[]> =>
    coreRequest.get(`${path}/search/${searchString}`),

  update: async (student: Student): Promise<string | undefined> =>
    (await coreRequest.put<{ id?: string }>(`${path}/${student.id}`, student)).id,
}
