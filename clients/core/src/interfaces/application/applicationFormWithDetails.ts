import type { ApplicationQuestionFileUpload } from '../../interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import type { ApplicationQuestionMultiSelect } from '../../interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import type { ApplicationQuestionText } from '../../interfaces/application/applicationQuestion/applicationQuestionText'
import type { OpenApplicationDetails } from '../../interfaces/application/openApplicationDetails'

export interface ApplicationFormWithDetails {
  applicationPhase: OpenApplicationDetails
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
}
