import { useState } from 'react'

import { AssessmentType } from '../../../interfaces/assessmentType'
import { useSchemaHasAssessmentData } from '../../hooks/useSchemaHasAssessmentData'
import {
  CoursePhaseConfig,
  CreateOrUpdateCoursePhaseConfigRequest,
} from '../../../interfaces/coursePhaseConfig'
import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import type { SchemaConfigurationCardProps } from '../components/SchemaConfigurationCard'
import { EvaluationOptions } from '../interfaces/EvaluationOption'
import {
  MainConfigState,
  useCoursePhaseConfigForm,
} from './useCoursePhaseConfigForm'
import { useCreateOrUpdateCoursePhaseConfig } from './useCreateOrUpdateCoursePhaseConfig'
import { useEvaluationOptions } from './useEvaluationOptions'

type EvaluationCardType = AssessmentType.SELF | AssessmentType.PEER | AssessmentType.TUTOR
type CardModel = Omit<SchemaConfigurationCardProps, 'schemas' | 'disabled' | 'isSaving'>

interface AssessmentVisibilityModel {
  gradingSheetVisible: boolean
  setGradingSheetVisible: (checked: boolean) => void
  gradeSuggestionVisible: boolean
  setGradeSuggestionVisible: (checked: boolean) => void
  actionItemsVisible: boolean
  setActionItemsVisible: (checked: boolean) => void
  evaluationResultsVisible: boolean
  setEvaluationResultsVisible: (checked: boolean) => void
}

export interface SettingsPageController {
  isSaving: boolean
  assessmentCard: CardModel
  evaluationCards: Record<EvaluationCardType, CardModel>
  assessmentVisibility: AssessmentVisibilityModel
}

const toDate = (value?: Date): Date | undefined => {
  if (!value) return undefined

  return new Date(value)
}

const areDatesEqual = (left?: Date, right?: Date) => left?.getTime() === right?.getTime()

const hasEvaluationCardChanges = (
  currentEnabled: boolean,
  currentSchema: string | undefined,
  currentStart: Date | undefined,
  currentDeadline: Date | undefined,
  originalEnabled: boolean | undefined,
  originalSchema: string | undefined,
  originalStart: Date | undefined,
  originalDeadline: Date | undefined,
) => {
  if (originalEnabled === undefined) {
    return Boolean(currentEnabled || currentSchema || currentStart || currentDeadline)
  }

  return (
    currentEnabled !== originalEnabled ||
    currentSchema !== (originalSchema || '') ||
    !areDatesEqual(currentStart, toDate(originalStart)) ||
    !areDatesEqual(currentDeadline, toDate(originalDeadline))
  )
}

const isEvaluationDetailReady = (
  currentEnabled: boolean,
  currentSchema: string | undefined,
  originalEnabled: boolean | undefined,
  originalSchema: string | undefined,
) =>
  currentEnabled &&
  currentSchema === (originalSchema ?? '') &&
  Boolean(originalEnabled) &&
  Boolean(originalSchema)

const buildRequestFromConfig = (
  config?: CoursePhaseConfig,
): CreateOrUpdateCoursePhaseConfigRequest => ({
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
})

const buildRequestFromDraft = (
  mainConfigState: MainConfigState,
  evaluationOptions: EvaluationOptions,
): CreateOrUpdateCoursePhaseConfigRequest => ({
  assessmentSchemaId: mainConfigState.assessmentSchemaId,
  start: mainConfigState.start,
  deadline: mainConfigState.deadline,
  selfEvaluationEnabled: evaluationOptions.self.enabled,
  selfEvaluationSchema: evaluationOptions.self.schema || undefined,
  selfEvaluationStart: evaluationOptions.self.start,
  selfEvaluationDeadline: evaluationOptions.self.deadline,
  peerEvaluationEnabled: evaluationOptions.peer.enabled,
  peerEvaluationSchema: evaluationOptions.peer.schema || undefined,
  peerEvaluationStart: evaluationOptions.peer.start,
  peerEvaluationDeadline: evaluationOptions.peer.deadline,
  tutorEvaluationEnabled: evaluationOptions.tutor.enabled,
  tutorEvaluationSchema: evaluationOptions.tutor.schema || undefined,
  tutorEvaluationStart: evaluationOptions.tutor.start,
  tutorEvaluationDeadline: evaluationOptions.tutor.deadline,
  evaluationResultsVisible: mainConfigState.evaluationResultsVisible,
  gradeSuggestionVisible: mainConfigState.gradeSuggestionVisible,
  actionItemsVisible: mainConfigState.actionItemsVisible,
  gradingSheetVisible: mainConfigState.gradingSheetVisible,
})

export const useSettingsPageController = (): SettingsPageController => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [activeErrorCard, setActiveErrorCard] = useState<AssessmentType | undefined>(undefined)
  const { coursePhaseConfig: originalConfig } = useCoursePhaseConfigStore()

  const {
    assessmentSchemaId,
    setAssessmentSchemaId,
    start,
    setStart,
    deadline,
    setDeadline,
    evaluationResultsVisible,
    setEvaluationResultsVisible,
    gradeSuggestionVisible,
    setGradeSuggestionVisible,
    actionItemsVisible,
    setActionItemsVisible,
    gradingSheetVisible,
    setGradingSheetVisible,
    mainConfigState,
    hasMainConfigChanges,
  } = useCoursePhaseConfigForm()

  const {
    selfEvaluationEnabled,
    setSelfEvaluationEnabled,
    selfEvaluationSchema,
    setSelfEvaluationSchema,
    selfEvaluationStart,
    setSelfEvaluationStart,
    selfEvaluationDeadline,
    setSelfEvaluationDeadline,
    peerEvaluationEnabled,
    setPeerEvaluationEnabled,
    peerEvaluationSchema,
    setPeerEvaluationSchema,
    peerEvaluationStart,
    setPeerEvaluationStart,
    peerEvaluationDeadline,
    setPeerEvaluationDeadline,
    tutorEvaluationEnabled,
    setTutorEvaluationEnabled,
    tutorEvaluationSchema,
    setTutorEvaluationSchema,
    tutorEvaluationStart,
    setTutorEvaluationStart,
    tutorEvaluationDeadline,
    setTutorEvaluationDeadline,
    evaluationOptions,
  } = useEvaluationOptions()

  const configMutation = useCreateOrUpdateCoursePhaseConfig(setError)

  const { data: assessmentSchemaData } = useSchemaHasAssessmentData(assessmentSchemaId || undefined)
  const { data: selfSchemaData } = useSchemaHasAssessmentData(
    selfEvaluationEnabled ? selfEvaluationSchema || undefined : undefined,
  )
  const { data: peerSchemaData } = useSchemaHasAssessmentData(
    peerEvaluationEnabled ? peerEvaluationSchema || undefined : undefined,
  )
  const { data: tutorSchemaData } = useSchemaHasAssessmentData(
    tutorEvaluationEnabled ? tutorEvaluationSchema || undefined : undefined,
  )

  const baseRequest = originalConfig
    ? buildRequestFromConfig(originalConfig)
    : buildRequestFromDraft(mainConfigState, evaluationOptions)

  const saveConfig = (
    assessmentType: AssessmentType,
    patch: Partial<CreateOrUpdateCoursePhaseConfigRequest>,
  ) => {
    setActiveErrorCard(assessmentType)
    configMutation.mutate({
      ...baseRequest,
      ...patch,
    })
  }

  const onCardError = (assessmentType: AssessmentType, nextError: string | undefined) => {
    setActiveErrorCard(assessmentType)
    setError(nextError)
  }

  const assessmentHasChanges = hasMainConfigChanges(originalConfig)
  const selfHasChanges = hasEvaluationCardChanges(
    selfEvaluationEnabled,
    selfEvaluationSchema,
    selfEvaluationStart,
    selfEvaluationDeadline,
    originalConfig?.selfEvaluationEnabled,
    originalConfig?.selfEvaluationSchema,
    originalConfig?.selfEvaluationStart,
    originalConfig?.selfEvaluationDeadline,
  )
  const peerHasChanges = hasEvaluationCardChanges(
    peerEvaluationEnabled,
    peerEvaluationSchema,
    peerEvaluationStart,
    peerEvaluationDeadline,
    originalConfig?.peerEvaluationEnabled,
    originalConfig?.peerEvaluationSchema,
    originalConfig?.peerEvaluationStart,
    originalConfig?.peerEvaluationDeadline,
  )
  const tutorHasChanges = hasEvaluationCardChanges(
    tutorEvaluationEnabled,
    tutorEvaluationSchema,
    tutorEvaluationStart,
    tutorEvaluationDeadline,
    originalConfig?.tutorEvaluationEnabled,
    originalConfig?.tutorEvaluationSchema,
    originalConfig?.tutorEvaluationStart,
    originalConfig?.tutorEvaluationDeadline,
  )

  const createDetailButtonLabel = (schemaId?: string) =>
    schemaId ? 'Save this card to open schema details' : 'Select a schema first'

  const assessmentCard: CardModel = {
    assessmentType: AssessmentType.ASSESSMENT,
    enabled: true,
    schemaId: assessmentSchemaId,
    onSchemaIdChange: setAssessmentSchemaId,
    startDate: start,
    onStartDateChange: setStart,
    deadline,
    onDeadlineChange: setDeadline,
    detailPath: assessmentSchemaId ? `schema/${assessmentSchemaId}` : '',
    canOpenDetails:
      assessmentSchemaId === (originalConfig?.assessmentSchemaID ?? '') &&
      Boolean(originalConfig?.assessmentSchemaID),
    detailButtonLabel: createDetailButtonLabel(assessmentSchemaId),
    hasAssessmentData: assessmentSchemaData?.hasAssessmentData ?? false,
    error: activeErrorCard === AssessmentType.ASSESSMENT ? error : undefined,
    hasChanges: assessmentHasChanges,
    onSave: () =>
      saveConfig(AssessmentType.ASSESSMENT, {
        assessmentSchemaId,
        start,
        deadline,
        evaluationResultsVisible,
        gradeSuggestionVisible,
        actionItemsVisible,
        gradingSheetVisible,
      }),
    canSave: Boolean(assessmentSchemaId),
    onCreateSchemaError: (nextError) => onCardError(AssessmentType.ASSESSMENT, nextError),
    showToggle: false,
  }

  const evaluationCards: Record<EvaluationCardType, CardModel> = {
    [AssessmentType.SELF]: {
      assessmentType: AssessmentType.SELF,
      enabled: selfEvaluationEnabled,
      onEnabledChange: setSelfEvaluationEnabled,
      schemaId: selfEvaluationSchema,
      onSchemaIdChange: setSelfEvaluationSchema,
      startDate: selfEvaluationStart,
      onStartDateChange: setSelfEvaluationStart,
      deadline: selfEvaluationDeadline,
      onDeadlineChange: setSelfEvaluationDeadline,
      detailPath: selfEvaluationSchema ? `schema/${selfEvaluationSchema}` : '',
      canOpenDetails: isEvaluationDetailReady(
        selfEvaluationEnabled,
        selfEvaluationSchema,
        originalConfig?.selfEvaluationEnabled,
        originalConfig?.selfEvaluationSchema,
      ),
      detailButtonLabel: createDetailButtonLabel(selfEvaluationSchema),
      hasAssessmentData: selfSchemaData?.hasAssessmentData ?? false,
      error: activeErrorCard === AssessmentType.SELF ? error : undefined,
      hasChanges: selfHasChanges,
      onSave: () =>
        saveConfig(AssessmentType.SELF, {
          selfEvaluationEnabled,
          selfEvaluationSchema: selfEvaluationSchema || undefined,
          selfEvaluationStart,
          selfEvaluationDeadline,
        }),
      canSave: !selfEvaluationEnabled || Boolean(selfEvaluationSchema),
      onCreateSchemaError: (nextError) => onCardError(AssessmentType.SELF, nextError),
    },
    [AssessmentType.PEER]: {
      assessmentType: AssessmentType.PEER,
      enabled: peerEvaluationEnabled,
      onEnabledChange: setPeerEvaluationEnabled,
      schemaId: peerEvaluationSchema,
      onSchemaIdChange: setPeerEvaluationSchema,
      startDate: peerEvaluationStart,
      onStartDateChange: setPeerEvaluationStart,
      deadline: peerEvaluationDeadline,
      onDeadlineChange: setPeerEvaluationDeadline,
      detailPath: peerEvaluationSchema ? `schema/${peerEvaluationSchema}` : '',
      canOpenDetails: isEvaluationDetailReady(
        peerEvaluationEnabled,
        peerEvaluationSchema,
        originalConfig?.peerEvaluationEnabled,
        originalConfig?.peerEvaluationSchema,
      ),
      detailButtonLabel: createDetailButtonLabel(peerEvaluationSchema),
      hasAssessmentData: peerSchemaData?.hasAssessmentData ?? false,
      error: activeErrorCard === AssessmentType.PEER ? error : undefined,
      hasChanges: peerHasChanges,
      onSave: () =>
        saveConfig(AssessmentType.PEER, {
          peerEvaluationEnabled,
          peerEvaluationSchema: peerEvaluationSchema || undefined,
          peerEvaluationStart,
          peerEvaluationDeadline,
        }),
      canSave: !peerEvaluationEnabled || Boolean(peerEvaluationSchema),
      onCreateSchemaError: (nextError) => onCardError(AssessmentType.PEER, nextError),
    },
    [AssessmentType.TUTOR]: {
      assessmentType: AssessmentType.TUTOR,
      enabled: tutorEvaluationEnabled,
      onEnabledChange: setTutorEvaluationEnabled,
      schemaId: tutorEvaluationSchema,
      onSchemaIdChange: setTutorEvaluationSchema,
      startDate: tutorEvaluationStart,
      onStartDateChange: setTutorEvaluationStart,
      deadline: tutorEvaluationDeadline,
      onDeadlineChange: setTutorEvaluationDeadline,
      detailPath: tutorEvaluationSchema ? `schema/${tutorEvaluationSchema}` : '',
      canOpenDetails: isEvaluationDetailReady(
        tutorEvaluationEnabled,
        tutorEvaluationSchema,
        originalConfig?.tutorEvaluationEnabled,
        originalConfig?.tutorEvaluationSchema,
      ),
      detailButtonLabel: createDetailButtonLabel(tutorEvaluationSchema),
      hasAssessmentData: tutorSchemaData?.hasAssessmentData ?? false,
      error: activeErrorCard === AssessmentType.TUTOR ? error : undefined,
      hasChanges: tutorHasChanges,
      onSave: () =>
        saveConfig(AssessmentType.TUTOR, {
          tutorEvaluationEnabled,
          tutorEvaluationSchema: tutorEvaluationSchema || undefined,
          tutorEvaluationStart,
          tutorEvaluationDeadline,
        }),
      canSave: !tutorEvaluationEnabled || Boolean(tutorEvaluationSchema),
      onCreateSchemaError: (nextError) => onCardError(AssessmentType.TUTOR, nextError),
    },
  }

  return {
    isSaving: configMutation.isPending,
    assessmentCard,
    evaluationCards,
    assessmentVisibility: {
      gradingSheetVisible,
      setGradingSheetVisible,
      gradeSuggestionVisible,
      setGradeSuggestionVisible,
      actionItemsVisible,
      setActionItemsVisible,
      evaluationResultsVisible,
      setEvaluationResultsVisible,
    },
  }
}
