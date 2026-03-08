import { Student } from '@tumaet/prompt-shared-state'
import { ApplicationAnswerMultiSelect } from './applicationAnswer/multiSelect/applicationAnswerMultiSelect'
import { ApplicationAnswerText } from './applicationAnswer/text/applicationAnswerText'
import { ApplicationAnswerFileUpload } from './applicationAnswer/fileUpload/applicationAnswerFileUpload'
import { ApplicationStatus } from './applicationStatus'

export interface GetApplication {
  id: string
  status: ApplicationStatus
  student?: Student
  answersText: ApplicationAnswerText[]
  answersMultiSelect: ApplicationAnswerMultiSelect[]
  answersFileUpload: ApplicationAnswerFileUpload[]
}
