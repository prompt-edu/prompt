import type {
  CreateFeedbackItemRequest,
  FeedbackItem,
  UpdateFeedbackItemRequest,
} from '../../interfaces/feedbackItem'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) =>
  `${coursePhasePath(coursePhaseID)}/evaluation/feedback-items`

export const feedbackItems = {
  listInPhase: (coursePhaseID: string): Promise<FeedbackItem[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  ofStudent: (coursePhaseID: string, courseParticipationID: string): Promise<FeedbackItem[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/course-participation/${courseParticipationID}`),

  ofTutor: (coursePhaseID: string, tutorParticipationID: string): Promise<FeedbackItem[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/tutor/${tutorParticipationID}`),

  listMine: (coursePhaseID: string): Promise<FeedbackItem[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-feedback`),

  create: (coursePhaseID: string, feedbackItem: CreateFeedbackItemRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), { ...feedbackItem, coursePhaseID }),

  update: (
    coursePhaseID: string,
    feedbackItemID: string,
    feedbackItem: UpdateFeedbackItemRequest,
  ): Promise<void> =>
    assessmentRequest.put(`${path(coursePhaseID)}/${feedbackItemID}`, {
      ...feedbackItem,
      coursePhaseID,
    }),

  remove: (coursePhaseID: string, feedbackItemID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${feedbackItemID}`),
}
