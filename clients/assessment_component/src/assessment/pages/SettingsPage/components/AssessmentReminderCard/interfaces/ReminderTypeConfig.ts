import type { EvaluationReminderType } from '../../../../../interfaces/evaluationReminder'

export interface ReminderTypeConfig {
  type: EvaluationReminderType
  label: string
  deadline?: Date
}
