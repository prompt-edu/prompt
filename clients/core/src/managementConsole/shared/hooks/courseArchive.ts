import type { CourseArchiveStatus } from '@core/interfaces/courseArchiveStatus'
import { coreApi } from '@core/network/api'
import { useCourseStore } from '@tumaet/prompt-shared-state'

/**
 * Archiving is the one write core mirrors into the course store rather than invalidating a cache,
 * so it lives here instead of in the request module: the endpoint returns the updated course, and
 * the sidebar reads its archived state from the store.
 */
const setArchived = async (courseId: string, payload: CourseArchiveStatus): Promise<void> => {
  const course = await coreApi.courses.setArchived(courseId, payload)
  const { updateCourse } = useCourseStore.getState()

  updateCourse(courseId, { archived: course.archived, archivedOn: course.archivedOn })
}

export const archiveCourses = async (courseIds: string[]): Promise<void> => {
  await Promise.all(courseIds.map((courseId) => setArchived(courseId, { archived: true })))
}

export const unarchiveCourses = async (courseIds: string[]): Promise<void> => {
  await Promise.all(courseIds.map((courseId) => setArchived(courseId, { archived: false })))
}
