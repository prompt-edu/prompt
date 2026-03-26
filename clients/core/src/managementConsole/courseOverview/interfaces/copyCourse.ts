export interface CopyCourse {
  name: string
  semesterTag: string
  startDate: Date
  endDate: Date
  template: boolean
  shortDescription?: string
  longDescription?: string
}

// Helper function to format Date as YYYY-MM-DD
function formatDate(date: Date): string {
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

// Serialize the CopyCourse object
export function serializeCopyCourse(course: CopyCourse): Record<string, any> {
  return {
    ...course,
    startDate: formatDate(course.startDate),
    endDate: formatDate(course.endDate),
  }
}
