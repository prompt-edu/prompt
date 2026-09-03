import type {
  AssessmentSchema,
  CreateAssessmentSchemaRequest,
  UpdateAssessmentSchemaRequest,
} from '../../interfaces/assessmentSchema'
import { assessmentRequest, coursePhasePath } from '../client'

export interface SchemaAssessmentDataResponse {
  hasAssessmentData: boolean
}

const path = (coursePhaseID: string) => `${coursePhasePath(coursePhaseID)}/assessment-schema`

export const schemas = {
  list: (coursePhaseID: string): Promise<AssessmentSchema[]> =>
    assessmentRequest.get(path(coursePhaseID)),

  hasAssessmentData: (
    coursePhaseID: string,
    schemaID: string,
  ): Promise<SchemaAssessmentDataResponse> =>
    assessmentRequest.get(`${path(coursePhaseID)}/${schemaID}/has-assessment-data`),

  create: (coursePhaseID: string, request: CreateAssessmentSchemaRequest): Promise<void> =>
    assessmentRequest.post(path(coursePhaseID), request),

  update: (
    coursePhaseID: string,
    schemaID: string,
    request: UpdateAssessmentSchemaRequest,
  ): Promise<void> => assessmentRequest.put(`${path(coursePhaseID)}/${schemaID}`, request),
}
