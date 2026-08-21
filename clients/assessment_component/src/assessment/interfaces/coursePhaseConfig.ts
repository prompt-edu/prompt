export interface CoursePhaseConfig {
  coursePhaseID: string // UUID as string
  assessmentEnabled: boolean
  assessmentSchemaID: string // UUID as string
  start: Date
  deadline: Date
  selfEvaluationEnabled: boolean
  selfEvaluationSchema: string // UUID as string
  selfEvaluationStart: Date
  selfEvaluationDeadline: Date
  peerEvaluationEnabled: boolean
  peerEvaluationSchema: string // UUID as string
  peerEvaluationStart: Date
  peerEvaluationDeadline: Date
  tutorEvaluationEnabled: boolean
  tutorEvaluationSchema: string // UUID as string
  tutorEvaluationStart: Date
  tutorEvaluationDeadline: Date
  evaluationResultsVisible: boolean
  gradeSuggestionVisible: boolean
  actionItemsVisible: boolean
  gradingSheetVisible: boolean
  resultsReleased: boolean
}

export interface CreateOrUpdateCoursePhaseConfigRequest {
  assessmentEnabled: boolean
  assessmentSchemaId: string
  start?: Date
  deadline?: Date
  selfEvaluationEnabled: boolean
  selfEvaluationSchema?: string
  selfEvaluationStart?: Date
  selfEvaluationDeadline?: Date
  peerEvaluationEnabled: boolean
  peerEvaluationSchema?: string
  peerEvaluationStart?: Date
  peerEvaluationDeadline?: Date
  tutorEvaluationEnabled: boolean
  tutorEvaluationSchema?: string
  tutorEvaluationStart?: Date
  tutorEvaluationDeadline?: Date
  evaluationResultsVisible: boolean
  gradeSuggestionVisible: boolean
  actionItemsVisible: boolean
  gradingSheetVisible: boolean
}
