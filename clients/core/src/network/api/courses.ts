import type { CourseArchiveStatus } from '@core/interfaces/courseArchiveStatus'
import type { CourseTemplateStatus } from '@core/interfaces/courseTemplateStatus'
import type { CheckCourseCopyableResponse } from '@core/managementConsole/courseOverview/interfaces/checkCourseCopyableResponse'
import {
  type CopyCourse,
  serializeCopyCourse,
} from '@core/managementConsole/courseOverview/interfaces/copyCourse'
import {
  type PostCourse,
  serializePostCourse,
  serializeUpdateCourse,
} from '@core/managementConsole/courseOverview/interfaces/postCourse'
import type { CourseParticipation } from '@core/managementConsole/shared/interfaces/CourseParticipation'
import type { Course, UpdateCourseData } from '@tumaet/prompt-shared-state'
import { isAxiosError } from 'axios'
import { API_PREFIX, coreRequest } from '../client'

const path = `${API_PREFIX}/courses`
const ofCourse = (courseID: string) => `${path}/${courseID}`

const UNAUTHORIZED = 401

/** A user with access to no course gets a 401 here rather than an empty list. */
const emptyOnUnauthorized = async <T>(read: Promise<T[]>): Promise<T[]> => {
  try {
    return await read
  } catch (error) {
    if (isAxiosError(error) && error.response?.status === UNAUTHORIZED) {
      return []
    }
    throw error
  }
}

export const courses = {
  list: (): Promise<Course[]> => emptyOnUnauthorized(coreRequest.get<Course[]>(`${path}/`)),

  listOwnIDs: (): Promise<string[]> =>
    emptyOnUnauthorized(coreRequest.get<string[]>(`${path}/self`)),

  listTemplates: (): Promise<Course[]> => coreRequest.get(`${path}/template`),

  nameExists: async (name: string, semesterTag: string): Promise<boolean> =>
    (
      await coreRequest.get<{ exists: boolean }>(`${path}/check-name`, {
        params: { name, semesterTag },
      })
    ).exists,

  myParticipation: (courseID: string): Promise<CourseParticipation> =>
    coreRequest.get(`${ofCourse(courseID)}/participations/self`),

  templateStatus: (courseID: string): Promise<CourseTemplateStatus> =>
    coreRequest.get(`${ofCourse(courseID)}/template`),

  copyability: (courseID: string): Promise<CheckCourseCopyableResponse> =>
    coreRequest.get(`${ofCourse(courseID)}/copyable`),

  create: async (course: PostCourse): Promise<string | undefined> =>
    (await coreRequest.post<{ id?: string }>(`${path}/`, serializePostCourse(course))).id,

  copy: async (courseID: string, courseVariables: CopyCourse): Promise<string | undefined> =>
    (
      await coreRequest.post<{ id?: string }>(
        `${ofCourse(courseID)}/copy`,
        serializeCopyCourse(courseVariables),
      )
    ).id,

  update: (courseID: string, courseData: UpdateCourseData): Promise<void> =>
    coreRequest.put(ofCourse(courseID), serializeUpdateCourse(courseData)),

  setTemplateStatus: (courseID: string, updateRequest: CourseTemplateStatus): Promise<void> =>
    coreRequest.put(`${ofCourse(courseID)}/template`, updateRequest),

  setArchived: (courseID: string, payload: CourseArchiveStatus): Promise<Course> =>
    coreRequest.put(`${ofCourse(courseID)}/archive`, payload),

  remove: (courseID: string): Promise<void> => coreRequest.del(ofCourse(courseID)),
}
