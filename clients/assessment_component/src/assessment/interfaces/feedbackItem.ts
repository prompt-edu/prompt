import type { AssessmentType } from './assessmentType'

export type FeedbackType = 'positive' | 'negative'

export interface FeedbackItem {
  id: string
  feedbackType: FeedbackType
  feedbackText: string
  courseParticipationID: string
  coursePhaseID: string
  authorCourseParticipationID: string
  createdAt: string
  type: AssessmentType
}

export interface CreateFeedbackItemRequest {
  feedbackType: FeedbackType
  feedbackText: string
  courseParticipationID: string
  authorCourseParticipationID: string
  type: AssessmentType
}

export interface UpdateFeedbackItemRequest {
  id: string
  feedbackType?: FeedbackType
  feedbackText?: string
  courseParticipationID?: string
  authorCourseParticipationID?: string
  type: AssessmentType
}
