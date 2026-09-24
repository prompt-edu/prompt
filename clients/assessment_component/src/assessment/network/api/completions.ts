import type {
  AssessmentCompletion,
  BatchMarkAssessmentCompletionsResult,
  BatchUnmarkAssessmentCompletionsResult,
  CreateOrUpdateAssessmentCompletionRequest,
} from '../../interfaces/assessmentCompletion'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) =>
  `${coursePhasePath(coursePhaseID)}/student-assessment/completed`

export const completions = {
  listInPhase: (coursePhaseID: string): Promise<AssessmentCompletion[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  myGradeSuggestion: (coursePhaseID: string): Promise<number | undefined> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-grade-suggestion`),

  save: (
    coursePhaseID: string,
    assessmentCompletion: CreateOrUpdateAssessmentCompletionRequest,
  ): Promise<void> => assessmentRequest.post(path(coursePhaseID), assessmentCompletion),

  markComplete: (
    coursePhaseID: string,
    assessmentCompletion: CreateOrUpdateAssessmentCompletionRequest,
  ): Promise<void> =>
    assessmentRequest.post(`${path(coursePhaseID)}/mark-complete`, assessmentCompletion),

  markCompleteBatch: (
    coursePhaseID: string,
    courseParticipationIDs: string[],
  ): Promise<BatchMarkAssessmentCompletionsResult> =>
    assessmentRequest.post(`${path(coursePhaseID)}/mark-complete/batch`, {
      courseParticipationIDs,
    }),

  unmarkBatch: (
    coursePhaseID: string,
    courseParticipationIDs: string[],
  ): Promise<BatchUnmarkAssessmentCompletionsResult> =>
    assessmentRequest.put(`${path(coursePhaseID)}/unmark/batch`, { courseParticipationIDs }),

  unmark: (coursePhaseID: string, courseParticipationID: string): Promise<void> =>
    assessmentRequest.put(
      `${path(coursePhaseID)}/course-participation/${courseParticipationID}/unmark`,
      {},
    ),

  remove: (coursePhaseID: string, courseParticipationID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/course-participation/${courseParticipationID}`),
}
