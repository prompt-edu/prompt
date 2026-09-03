import { axiosInstance, type Team } from '@tumaet/prompt-shared-state'

import type { AssessmentParticipationWithStudent } from '../../interfaces/assessmentParticipationWithStudent'
import type {
  CoursePhaseConfig,
  CreateOrUpdateCoursePhaseConfigRequest,
} from '../../interfaces/coursePhaseConfig'
import type {
  EvaluationReminderReport,
  SendEvaluationReminderRequest,
} from '../../interfaces/evaluationReminder'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/config`

export const config = {
  get: (coursePhaseID: string): Promise<CoursePhaseConfig> =>
    assessmentRequest.get(path(coursePhaseID)),

  participations: (coursePhaseID: string): Promise<AssessmentParticipationWithStudent[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/participations`),

  teams: (coursePhaseID: string): Promise<Team[]> =>
    assessmentRequest.get(`${path(coursePhaseID)}/teams`),

  save: (coursePhaseID: string, request: CreateOrUpdateCoursePhaseConfigRequest): Promise<void> =>
    assessmentRequest.put(path(coursePhaseID), request),

  releaseResults: (coursePhaseID: string): Promise<void> =>
    assessmentRequest.post(`${path(coursePhaseID)}/release`, {}),

  unreleaseResults: (coursePhaseID: string): Promise<void> =>
    assessmentRequest.post(`${path(coursePhaseID)}/unrelease`, {}),

  // The only endpoint on core's host rather than the assessment host: in production both resolve
  // to the same origin, where traefik routes /assessment/api to this service
  sendReminder: async (
    coursePhaseID: string,
    request: SendEvaluationReminderRequest,
  ): Promise<EvaluationReminderReport> =>
    (
      await axiosInstance.post<EvaluationReminderReport>(
        `/${path(coursePhaseID)}/reminders/send`,
        request,
        { headers: { 'Content-Type': 'application/json' } },
      )
    ).data,
}
