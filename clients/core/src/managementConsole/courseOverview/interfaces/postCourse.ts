import type { UpdateCourseData } from '@tumaet/prompt-shared-state'

export interface PostCourse {
  name: string
  startDate: Date
  endDate: Date
  courseType: string
  ects: number
  semesterTag: string
  restrictedMetaData: { [key: string]: any }
  studentReadableData: { [key: string]: any }
  shortDescription: string
  longDescription?: string
  template: boolean
  orgID?: string
}

// Helper function to format Date as YYYY-MM-DD
function formatDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Serialize the PostCourse object
export function serializePostCourse(course: PostCourse): Record<string, any> {
  return {
    ...course,
    startDate: formatDate(course.startDate),
    endDate: formatDate(course.endDate),
  }
}

export function serializeUpdateCourse(course: UpdateCourseData): Record<string, any> {
  return {
    ...course,
    startDate: course.startDate ? formatDate(course.startDate) : undefined,
    endDate: course.endDate ? formatDate(course.endDate) : undefined,
    shortDescription: course.shortDescription ?? undefined,
    longDescription: course.longDescription ?? undefined,
  }
}
