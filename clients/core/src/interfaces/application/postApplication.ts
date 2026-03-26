import { CreateApplicationAnswerMultiSelect } from '../application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'
import { CreateApplicationAnswerText } from '../application/applicationAnswer/text/createApplicationAnswerText'
import { CreateApplicationAnswerFileUpload } from '../application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { Student } from '@tumaet/prompt-shared-state'

export interface PostApplication {
  student: Student
  answersText: CreateApplicationAnswerText[]
  answersMultiSelect: CreateApplicationAnswerMultiSelect[]
  answersFileUpload: CreateApplicationAnswerFileUpload[]
}
