import type { Student } from '@tumaet/prompt-shared-state'
import type { CreateApplicationAnswerFileUpload } from '../application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import type { CreateApplicationAnswerMultiSelect } from '../application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import type { CreateApplicationAnswerText } from '../application/applicationAnswer/text/createApplicationAnswerText'

export interface PostApplication {
  student: Student
  answersText: CreateApplicationAnswerText[]
  answersMultiSelect: CreateApplicationAnswerMultiSelect[]
  answersFileUpload: CreateApplicationAnswerFileUpload[]
}
