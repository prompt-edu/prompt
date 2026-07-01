import { ApplicationQuestionFileUpload } from '../../interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { ApplicationQuestionMultiSelect } from '../../interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '../../interfaces/application/applicationQuestion/applicationQuestionText'
import { OpenApplicationDetails } from '../../interfaces/application/openApplicationDetails'

export interface ApplicationFormWithDetails {
  applicationPhase: OpenApplicationDetails
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
}
