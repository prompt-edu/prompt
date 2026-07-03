import type { CreateApplicationAnswerMultiSelect } from '@core/interfaces/application/applicationAnswer/multiSelect/createApplicationAnswerMultiSelect'

export interface QuestionMultiSelectFormRef {
  validate: () => Promise<boolean>
  getValues: () => CreateApplicationAnswerMultiSelect
}
