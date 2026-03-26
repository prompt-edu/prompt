import { ApplicationQuestionMultiSelect } from '../../interfaces/application/applicationQuestion/applicationQuestionMultiSelect'
import { ApplicationQuestionText } from '../../interfaces/application/applicationQuestion/applicationQuestionText'
import { ApplicationQuestionFileUpload } from '../../interfaces/application/applicationQuestion/applicationQuestionFileUpload'
import { OpenApplicationDetails } from '../../interfaces/application/openApplicationDetails'

export interface ApplicationFormWithDetails {
  applicationPhase: OpenApplicationDetails
  questionsText: ApplicationQuestionText[]
  questionsMultiSelect: ApplicationQuestionMultiSelect[]
  questionsFileUpload: ApplicationQuestionFileUpload[]
}
