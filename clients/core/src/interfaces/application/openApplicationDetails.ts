import type { CourseType } from '@tumaet/prompt-shared-state'
import type { ProfilePictureRequirement } from '@tumaet/prompt-ui-components'

/** Whether and why the application asks logged-in applicants for a profile picture. */
export interface ApplicationProfilePictureConfig {
  requirement: ProfilePictureRequirement
  explanation?: string
  /** Reviewers do not see the picture until the applicant is accepted. */
  hiddenUntilAccepted: boolean
}

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
  welcomeText?: string | null
  /** Only set on the form of one application. */
  profilePicture?: ApplicationProfilePictureConfig
}
