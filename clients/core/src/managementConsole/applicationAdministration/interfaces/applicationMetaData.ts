import type { ProfilePictureRequirement } from '@tumaet/prompt-ui-components'

export type ApplicationMode = 'apply' | 'import'

export type ApplicationMetaData = {
  applicationStartDate?: Date
  applicationEndDate?: Date
  externalStudentsAllowed?: boolean
  universityLoginAvailable?: boolean
  autoAccept?: boolean
  useCustomScores?: boolean
  applicationCsvExportSettings?: Record<string, boolean>
  applicationMode?: ApplicationMode
  welcomeText?: string
  profilePictureRequirement?: ProfilePictureRequirement
  profilePictureExplanation?: string
  hideProfilePicturesUntilAccepted?: boolean
}
