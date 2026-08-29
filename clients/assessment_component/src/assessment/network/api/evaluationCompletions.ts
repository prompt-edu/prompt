import type {
  EvaluationCompletion,
  EvaluationCompletionRequest,
} from '../../interfaces/evaluationCompletion'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/evaluation/completed`

export const evaluationCompletions = {
  listInPhase: (coursePhaseID: string): Promise<EvaluationCompletion[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  listMine: (coursePhaseID: string): Promise<EvaluationCompletion[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-completions`),

  markMine: (
    coursePhaseID: string,
    evaluationCompletion: EvaluationCompletionRequest,
  ): Promise<void> =>
    assessmentRequest.post(
      `${path(coursePhaseID)}/my-completion/mark-complete`,
      evaluationCompletion,
    ),

  unmarkMine: (
    coursePhaseID: string,
    evaluationCompletion: EvaluationCompletionRequest,
  ): Promise<void> =>
    assessmentRequest.put(`${path(coursePhaseID)}/my-completion/unmark`, evaluationCompletion),
}
