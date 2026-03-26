import { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'

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
