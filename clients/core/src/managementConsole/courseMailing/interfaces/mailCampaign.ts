export enum MailCampaignStatus {
  Draft = 'draft',
  Sending = 'sending',
  Sent = 'sent',
  PartiallyFailed = 'partially_failed',
  Failed = 'failed',
}

export enum RecipientStatus {
  Pending = 'pending',
  Sent = 'sent',
  Failed = 'failed',
}

export interface MailItem {
  name: string
  email: string
}

export interface Actor {
  id: string
  email: string
  name: string
}

export interface MailCampaign {
  id: string
  courseID: string
  name: string
  subject: string
  body: string
  targetCoursePhaseID: string | null
  targetPassStatuses: string[]
  replyToOverride: MailItem | null
  ccOverride: MailItem[] | null
  bccOverride: MailItem[] | null
  status: MailCampaignStatus
  createdAt: string
  createdBy: Actor
  updatedAt: string
  updatedBy: Actor
  sentAt: string | null
  sentBy: Actor | null
  recipientCount: number
  sentCount: number
  failedCount: number
  pendingCount: number
}

export interface MailCampaignRecipient {
  courseParticipationID: string
  firstName: string
  lastName: string
  email: string
  status: RecipientStatus
  errorMessage: string
  sentAt: string | null
}

export interface MailCampaignDetail extends MailCampaign {
  recipients: MailCampaignRecipient[]
}

export interface MailCampaignRequest {
  name: string
  subject: string
  body: string
  targetCoursePhaseID: string | null
  targetPassStatuses: string[]
  replyToOverride: MailItem | null
  ccOverride: MailItem[]
  bccOverride: MailItem[]
}

export interface RecipientPreviewItem {
  courseParticipationID: string
  firstName: string
  lastName: string
  email: string
}

export interface RecipientPreview {
  count: number
  recipients: RecipientPreviewItem[]
}

export interface SendResponse {
  recipientCount: number
}
