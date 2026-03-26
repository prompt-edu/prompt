import { AssessmentType } from '../../../interfaces/assessmentType'
import {
  CoursePhaseConfig,
  CreateOrUpdateCoursePhaseConfigRequest,
} from '../../../interfaces/coursePhaseConfig'

export type EvaluationAssessmentType =
  | AssessmentType.SELF
  | AssessmentType.PEER
  | AssessmentType.TUTOR

interface EvaluationStateSnapshot {
  enabled: boolean
  schema: string
  start?: Date
  deadline?: Date
}

interface EvaluationOriginalSnapshot {
  enabled?: boolean
  schema?: string
  start?: Date
  deadline?: Date
}

interface AssessmentStateSnapshot {
  assessmentSchemaId: string
  start?: Date
  deadline?: Date
  evaluationResultsVisible: boolean
  gradeSuggestionVisible: boolean
  actionItemsVisible: boolean
  gradingSheetVisible: boolean
}

const toDate = (value?: Date | string): Date | undefined => {
  if (!value) return undefined

  return new Date(value)
}

const areDatesEqual = (left?: Date, right?: Date) => left?.getTime() === right?.getTime()

export const createDetailButtonLabel = (schemaId?: string) =>
  schemaId ? 'Save this card to open schema details' : 'Select a schema first'

export const buildRequestFromConfig = (
  config?: CoursePhaseConfig,
): CreateOrUpdateCoursePhaseConfigRequest | undefined => {
  if (!config) {
    return undefined
  }

  return {
    assessmentSchemaId: config?.assessmentSchemaID ?? '',
    start: toDate(config?.start),
    deadline: toDate(config?.deadline),
    selfEvaluationEnabled: config?.selfEvaluationEnabled ?? false,
    selfEvaluationSchema: config?.selfEvaluationSchema || undefined,
    selfEvaluationStart: toDate(config?.selfEvaluationStart),
    selfEvaluationDeadline: toDate(config?.selfEvaluationDeadline),
    peerEvaluationEnabled: config?.peerEvaluationEnabled ?? false,
    peerEvaluationSchema: config?.peerEvaluationSchema || undefined,
    peerEvaluationStart: toDate(config?.peerEvaluationStart),
    peerEvaluationDeadline: toDate(config?.peerEvaluationDeadline),
    tutorEvaluationEnabled: config?.tutorEvaluationEnabled ?? false,
    tutorEvaluationSchema: config?.tutorEvaluationSchema || undefined,
    tutorEvaluationStart: toDate(config?.tutorEvaluationStart),
    tutorEvaluationDeadline: toDate(config?.tutorEvaluationDeadline),
    evaluationResultsVisible: config?.evaluationResultsVisible ?? false,
    gradeSuggestionVisible: config?.gradeSuggestionVisible ?? true,
    actionItemsVisible: config?.actionItemsVisible ?? true,
    gradingSheetVisible: config?.gradingSheetVisible ?? false,
  }
}

export const hasAssessmentCardChanges = (
  current: AssessmentStateSnapshot,
  originalConfig?: CoursePhaseConfig,
) => {
  if (!originalConfig) return true

  return (
    current.assessmentSchemaId !== (originalConfig.assessmentSchemaID || '') ||
    !areDatesEqual(current.start, toDate(originalConfig.start)) ||
    !areDatesEqual(current.deadline, toDate(originalConfig.deadline)) ||
    current.evaluationResultsVisible !== (originalConfig.evaluationResultsVisible || false) ||
    current.gradeSuggestionVisible !== (originalConfig.gradeSuggestionVisible ?? true) ||
    current.actionItemsVisible !== (originalConfig.actionItemsVisible ?? true) ||
    current.gradingSheetVisible !== (originalConfig.gradingSheetVisible ?? false)
  )
}

export const getEvaluationOriginalSnapshot = (
  config: CoursePhaseConfig | undefined,
  assessmentType: EvaluationAssessmentType,
): EvaluationOriginalSnapshot => {
  switch (assessmentType) {
    case AssessmentType.SELF:
      return {
        enabled: config?.selfEvaluationEnabled,
        schema: config?.selfEvaluationSchema,
        start: toDate(config?.selfEvaluationStart),
        deadline: toDate(config?.selfEvaluationDeadline),
      }
    case AssessmentType.PEER:
      return {
        enabled: config?.peerEvaluationEnabled,
        schema: config?.peerEvaluationSchema,
        start: toDate(config?.peerEvaluationStart),
        deadline: toDate(config?.peerEvaluationDeadline),
      }
    case AssessmentType.TUTOR:
      return {
        enabled: config?.tutorEvaluationEnabled,
        schema: config?.tutorEvaluationSchema,
        start: toDate(config?.tutorEvaluationStart),
        deadline: toDate(config?.tutorEvaluationDeadline),
      }
  }
}

export const hasEvaluationCardChanges = (
  current: EvaluationStateSnapshot,
  original: EvaluationOriginalSnapshot,
) => {
  if (original.enabled === undefined) {
    return Boolean(current.enabled || current.schema || current.start || current.deadline)
  }

  return (
    current.enabled !== original.enabled ||
    current.schema !== (original.schema || '') ||
    !areDatesEqual(current.start, original.start) ||
    !areDatesEqual(current.deadline, original.deadline)
  )
}

export const isEvaluationDetailReady = (
  currentEnabled: boolean,
  currentSchema: string | undefined,
  originalEnabled: boolean | undefined,
  originalSchema: string | undefined,
) =>
  currentEnabled &&
  currentSchema === (originalSchema ?? '') &&
  Boolean(originalEnabled) &&
  Boolean(originalSchema)

export const buildEvaluationPatch = (
  assessmentType: EvaluationAssessmentType,
  state: EvaluationStateSnapshot,
): Partial<CreateOrUpdateCoursePhaseConfigRequest> => {
  switch (assessmentType) {
    case AssessmentType.SELF:
      return {
        selfEvaluationEnabled: state.enabled,
        selfEvaluationSchema: state.schema || undefined,
        selfEvaluationStart: state.start,
        selfEvaluationDeadline: state.deadline,
      }
    case AssessmentType.PEER:
      return {
        peerEvaluationEnabled: state.enabled,
        peerEvaluationSchema: state.schema || undefined,
        peerEvaluationStart: state.start,
        peerEvaluationDeadline: state.deadline,
      }
    case AssessmentType.TUTOR:
      return {
        tutorEvaluationEnabled: state.enabled,
        tutorEvaluationSchema: state.schema || undefined,
        tutorEvaluationStart: state.start,
        tutorEvaluationDeadline: state.deadline,
      }
  }
}
