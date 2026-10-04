import type { EvaluationReminderType } from '../../../../../interfaces/evaluationReminder'

export interface ReminderTypeConfig {
  type: EvaluationReminderType
  label: string
  // The label as it reads mid-sentence; only the built-in words are lowercased.
  inlineLabel: string
  deadline?: Date
}
