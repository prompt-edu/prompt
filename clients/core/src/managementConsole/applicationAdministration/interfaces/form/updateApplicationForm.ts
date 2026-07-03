import type { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'

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
