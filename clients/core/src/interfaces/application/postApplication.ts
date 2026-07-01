import { Student } from '@tumaet/prompt-shared-state'
import { CreateApplicationAnswerFileUpload } from '../application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { CreateApplicationAnswerMultiSelect } from '../application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import { CreateApplicationAnswerText } from '../application/applicationAnswer/text/createApplicationAnswerText'

export interface PostApplication {
  student: Student
  answersText: CreateApplicationAnswerText[]
  answersMultiSelect: CreateApplicationAnswerMultiSelect[]
  answersFileUpload: CreateApplicationAnswerFileUpload[]
}
