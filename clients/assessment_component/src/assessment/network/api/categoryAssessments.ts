import type { CreateOrUpdateCategoryAssessmentRequest } from '../../interfaces/categoryAssessment'
import { assessmentRequest, coursePhasePath, WRITE_TIMEOUT_MS } from '../client'

export const categoryAssessments = {
  save: (
    coursePhaseID: string,
    categoryAssessment: CreateOrUpdateCategoryAssessmentRequest,
  ): Promise<void> =>
    assessmentRequest.post(
      `${coursePhasePath(coursePhaseID)}/category-assessment`,
      categoryAssessment,
      { timeoutMs: WRITE_TIMEOUT_MS },
    ),
}
