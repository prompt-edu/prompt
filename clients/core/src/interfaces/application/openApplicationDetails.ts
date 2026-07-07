import type { CourseType } from '@tumaet/prompt-shared-state'

export interface OpenApplicationDetails {
  id: string
  courseName: string
  courseType: CourseType
  ects: number
  startDate: Date
  endDate: Date
  applicationDeadline: Date
  externalStudentsAllowed: boolean
  universityLoginAvailable: boolean
  shortDescription?: string | null
  longDescription?: string | null
}
