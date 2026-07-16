export type TargetMode = 'individual' | 'team'
export type FeedbackMode = 'independent' | 'shared'
export type FeedbackStatus = 'draft' | 'submitted'
export type TargetType = 'individual' | 'team'

export interface PresentationConfig {
  coursePhaseId: string
  targetMode: TargetMode
  feedbackMode: FeedbackMode
}

export interface FeedbackCategory {
  id: string
  coursePhaseId?: string
  name: string
  description: string
  position: number
}

export interface PresentationSlot {
  id: string
  coursePhaseId?: string
  startTime: string
  endTime: string
  location?: string
  assigned?: boolean
  presentation?: PresentationSummary
}

export interface PresentationSummary {
  id: string
  coursePhaseId: string
  slotId: string
  targetType: TargetType
  targetId: string
  targetName: string
  startTime: string
  endTime: string
  location?: string
  materialCount?: number
  feedbackCount?: number
  submittedFeedbackCount?: number
  feedbackReleasedAt?: string
  feedbackReleasedByName?: string
  feedbackReleaseName?: string
}

export interface PresentationTarget {
  id: string
  name: string
  type: TargetType
  assigned?: boolean
  assignedPresentationId?: string
}

export interface PresentationMaterial {
  id: string
  presentationId?: string
  fileName: string
  contentType: string
  sizeBytes: number
  uploadedByName?: string
  uploadedAt: string
}

export interface MaterialUploadIntent {
  uploadId: string
  uploadUrl: string
  headers?: Record<string, string>
  expiresAt: string
}

export interface MaterialDownload {
  downloadUrl: string
  fileName: string
  expiresAt: string
}

export interface FeedbackAnswer {
  categoryId: string
  value: string
  revision: number
  updatedByName?: string
  updatedAt?: string
}

export interface FeedbackContributor {
  userId: string
  name: string
  firstContributedAt?: string
  lastContributedAt?: string
}

export interface FeedbackForm {
  id: string
  evaluatorName?: string
  status: FeedbackStatus
  submittedAt?: string
  answers: FeedbackAnswer[]
  contributors: FeedbackContributor[]
  isOwn: boolean
}

export interface FeedbackDocument {
  presentation: PresentationSummary
  mode: FeedbackMode
  categories: FeedbackCategory[]
  released?: boolean
  ownForm?: FeedbackForm
  forms: FeedbackForm[]
  activeEditors: ActiveEditor[]
  canEdit: boolean
  canRelease: boolean
}

export interface ActiveEditor {
  connectionId: string
  userId: string
  name: string
  expiresAt: string
}

export interface FeedbackConflict {
  code: 'feedback_conflict'
  message: string
  answer: FeedbackAnswer
}

export interface FeedbackEvent {
  id?: string
  type: 'snapshot' | 'answer.updated' | 'form.status.changed' | 'presence.updated' | 'released'
  presentationId: string
  answer?: FeedbackAnswer
  activeEditors?: ActiveEditor[]
}

export interface CreateSlotRequest {
  startTime: string
  endTime: string
  location?: string
}

export interface CategoryRequest {
  name: string
  description: string
  position: number
}

export interface PresentationApiError {
  code?: string
  error?: string
  message?: string
  details?: unknown
  answer?: FeedbackAnswer
}
