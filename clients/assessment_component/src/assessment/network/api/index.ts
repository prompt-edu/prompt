import { actionItems } from './actionItems'
import { assessments } from './assessments'
import { categories } from './categories'
import { categoryAssessments } from './categoryAssessments'
import { competencies } from './competencies'
import { completions } from './completions'
import { config } from './config'
import { evaluationCompletions } from './evaluationCompletions'
import { evaluations } from './evaluations'
import { feedbackItems } from './feedbackItems'
import { schemas } from './schemas'

export const assessmentApi = {
  actionItems,
  assessments,
  categories,
  categoryAssessments,
  competencies,
  completions,
  config,
  evaluationCompletions,
  evaluations,
  feedbackItems,
  schemas,
}

export type { AssessmentExportFormat } from './assessments'
export type { SchemaAssessmentDataResponse } from './schemas'
