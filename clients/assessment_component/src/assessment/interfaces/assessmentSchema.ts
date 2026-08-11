export interface AssessmentSchema {
  id: string
  name: string
  description: string
  createdAt?: Date
  updatedAt?: Date
  isOwnedByCurrentPhase?: boolean
}

export interface CreateAssessmentSchemaRequest {
  name: string
  description: string
}

export interface UpdateAssessmentSchemaRequest {
  name: string
  description: string
}

export interface CreateOrUpdateAssessmentSchemaCoursePhaseRequest {
  assessmentSchemaID: string
  coursePhaseID: string
}
