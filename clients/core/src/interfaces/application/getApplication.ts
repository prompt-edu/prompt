import type { Student } from '@tumaet/prompt-shared-state'
import type { ApplicationAnswerFileUpload } from './applicationAnswer/fileUpload/applicationAnswerFileUpload'
import type { ApplicationAnswerMultiSelect } from './applicationAnswer/multiSelect/applicationAnswerMultiSelect'
import type { ApplicationAnswerText } from './applicationAnswer/text/applicationAnswerText'
import type { ApplicationStatus } from './applicationStatus'

export interface GetApplication {
  id: string
  status: ApplicationStatus
  student?: Student
  answersText: ApplicationAnswerText[]
  answersMultiSelect: ApplicationAnswerMultiSelect[]
  answersFileUpload: ApplicationAnswerFileUpload[]
}
