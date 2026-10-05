export interface ResultsReleasedMailSettings {
  subject: string
  content: string
  sendOnRelease: boolean
}

export interface ResultsReleasedMailReport {
  successfulEmails: string[]
  failedEmails: string[]
  requestedRecipients: number
}

export interface ReleaseResultsResponse {
  message: string
  mailReport?: ResultsReleasedMailReport
  mailError?: string
}
