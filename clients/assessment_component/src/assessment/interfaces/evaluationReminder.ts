export type EvaluationReminderType = 'self' | 'peer' | 'tutor'

export interface AssessmentReminderMetaData {
  subject: string
  content: string
  lastSentAtByType: Partial<Record<EvaluationReminderType, string>>
}

export interface SendEvaluationReminderRequest {
  evaluationType: EvaluationReminderType
}

export interface EvaluationReminderReport {
  successfulEmails: string[]
  failedEmails: string[]
  requestedRecipients: number
  evaluationType: EvaluationReminderType
  deadline?: string
  deadlinePassed: boolean
  sentAt: string
  previousSentAt?: string | null
}
