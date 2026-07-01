import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'

export interface UpdateApplicationForm {
  deleteQuestionsText: string[]
  deleteQuestionsMultiSelect: string[]
  deleteQuestionsFileUpload: string[]
  createQuestionsText: ApplicationQuestionText[]
  createQuestionsMultiSelect: ApplicationQuestionMultiSelect[]
  createQuestionsFileUpload: ApplicationQuestionFileUpload[]
  updateQuestionsText: ApplicationQuestionText[]
  updateQuestionsMultiSelect: ApplicationQuestionMultiSelect[]
  updateQuestionsFileUpload: ApplicationQuestionFileUpload[]
}
