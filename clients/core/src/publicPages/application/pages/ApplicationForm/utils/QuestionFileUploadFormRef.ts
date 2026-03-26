import { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'
import { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'

export interface QuestionFileUploadFormRef {
  validate: () => Promise<boolean>
  getValues: () => CreateApplicationAnswerFileUpload
  rerender: (answer?: ApplicationAnswerFileUpload) => void
}
