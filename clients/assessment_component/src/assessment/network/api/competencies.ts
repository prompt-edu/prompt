import type { CreateCompetencyRequest, UpdateCompetencyRequest } from '../../interfaces/competency'
import { assessmentRequest, coursePhasePath } from '../client'

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/competency`

export const competencies = {
  create: (coursePhaseID: string, competency: CreateCompetencyRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), competency),

  update: (coursePhaseID: string, competency: UpdateCompetencyRequest): Promise<void> =>
    assessmentRequest.put(`${path(coursePhaseID)}/${competency.id}`, competency),

  remove: (coursePhaseID: string, competencyID: string): Promise<void> =>
    assessmentRequest.del(`${path(coursePhaseID)}/${competencyID}`),
}
