import type { ApplicationQuestionFileUpload } from '@core/interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '@core/interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '@core/interfaces/application/applicationQuestion/applicationQuestionText'

export interface ApplicationForm {
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
}
