import type { Assessment, CreateOrUpdateAssessmentRequest } from '../../interfaces/assessment'
import type { StudentAssessmentResults } from '../../interfaces/assessmentResults'
import type { ScoreLevelWithParticipation } from '../../interfaces/scoreLevelWithParticipation'
import type { StudentAssessment } from '../../interfaces/studentAssessment'
import { assessmentRequest, coursePhasePath, WRITE_TIMEOUT_MS } from '../client'

export type AssessmentExportFormat = 'json'

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/student-assessment`

export const assessments = {
  listInPhase: (coursePhaseID: string): Promise<Assessment[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  ofParticipant: (
    coursePhaseID: string,
    courseParticipationID: string,
  ): Promise<StudentAssessment> =>
    assessmentRequest.get(`${path(coursePhaseID)}/${courseParticipationID}`),

  scoreLevels: (coursePhaseID: string): Promise<ScoreLevelWithParticipation[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/scoreLevel`),

  myResults: (coursePhaseID: string): Promise<StudentAssessmentResults> =>
    assessmentRequest.get(`${path(coursePhaseID)}/my-results`),

  export: (
    coursePhaseID: string,
    courseParticipationID: string,
    format: AssessmentExportFormat,
  ): Promise<unknown> =>
    assessmentRequest.get(`${path(coursePhaseID)}/${courseParticipationID}/export`, { format }),

  save: (coursePhaseID: string, assessment: CreateOrUpdateAssessmentRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), assessment, { timeoutMs: WRITE_TIMEOUT_MS }),

  remove: (coursePhaseID: string, assessmentID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${assessmentID}`),
}
