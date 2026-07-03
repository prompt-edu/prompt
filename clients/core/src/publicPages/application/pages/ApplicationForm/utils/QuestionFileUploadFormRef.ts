import type { ApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/applicationAnswerFileUpload'
import type { CreateApplicationAnswerFileUpload } from '@core/interfaces/application/applicationAnswer/fileUpload/createApplicationAnswerFileUpload'

export interface QuestionFileUploadFormRef {
  validate: () => Promise<boolean>
  getValues: () => CreateApplicationAnswerFileUpload
  rerender: (answer?: ApplicationAnswerFileUpload) => void
}
