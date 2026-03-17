import { axiosInstance } from '@/network/configService'
import type { CourseArchiveStatus } from '@core/interfaces/courseArchiveStatus'
import { useCourseStore } from '@tumaet/prompt-shared-state'
import type { Course } from '@tumaet/prompt-shared-state'

const updateCourseArchiveStatus = async (
  courseId: string,
  payload: CourseArchiveStatus,
): Promise<void> => {
  try {
    const response = await axiosInstance.put<Course>(`/api/courses/${courseId}/archive`, payload, {
      headers: {
        'Content-Type': 'application/json',
      },
    })

    const { updateCourse } = useCourseStore.getState()

    updateCourse(courseId, {
      archived: response.data.archived,
      archivedOn: response.data.archivedOn,
    })
  } catch (err) {
    console.error('Failed to update course archive status', err)
    throw err
  }
}

export const archiveCourses = async (courseIds: string[]): Promise<void> => {
  await Promise.all(
    courseIds.map((courseId) => updateCourseArchiveStatus(courseId, { archived: true })),
  )
}

export const unarchiveCourses = async (courseIds: string[]): Promise<void> => {
  await Promise.all(
    courseIds.map((courseId) => updateCourseArchiveStatus(courseId, { archived: false })),
  )
}
