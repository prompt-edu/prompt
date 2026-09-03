import type { CreateOrUpdateEvaluationRequest, Evaluation } from '../../interfaces/evaluation'
import type { StudentEvaluationResults } from '../../interfaces/evaluationResults'
import { assessmentRequest, coursePhasePath } from '../client'

const NO_CONTENT = 204

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/evaluation`

export const evaluations = {
  listInPhase: (coursePhaseID: string): Promise<Evaluation[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  listMine: (coursePhaseID: string): Promise<Evaluation[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-evaluations`),

  ofSelf: (coursePhaseID: string, courseParticipationID: string): Promise<Evaluation[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/self/${courseParticipationID}`),

  ofPeers: (coursePhaseID: string, courseParticipationID: string): Promise<Evaluation[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/peer/${courseParticipationID}`),

  ofTutor: (coursePhaseID: string, tutorParticipationID: string): Promise<Evaluation[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/tutor/${tutorParticipationID}`),

  // 204 when the phase is not evaluation-only or results are not released yet.
  // React Query rejects an undefined result, so this must be null.
  myResults: async (coursePhaseID: string): Promise<StudentEvaluationResults | null> => {
    const response = await assessmentRequest.getResponse<StudentEvaluationResults>(
      `${path(coursePhaseID)}/my-results`,
    )
    return response.status === NO_CONTENT ? null : response.data
  },

  save: (coursePhaseID: string, evaluation: CreateOrUpdateEvaluationRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), evaluation),

  remove: (coursePhaseID: string, evaluationID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${evaluationID}`),
}
