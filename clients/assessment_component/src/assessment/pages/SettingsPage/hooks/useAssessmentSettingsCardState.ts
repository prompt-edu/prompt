import { useEffect, useMemo, useState } from 'react'

import { AssessmentType } from '../../../interfaces/assessmentType'
import { useGetCoursePhaseConfig } from '../../hooks/useGetCoursePhaseConfig'
import { useSchemaHasAssessmentData } from '../../hooks/useSchemaHasAssessmentData'
import type { SchemaConfigurationCardProps } from '../components/SchemaConfigurationCard'
import {
  buildRequestFromConfig,
  createDetailButtonLabel,
  hasAssessmentCardChanges,
} from './settingsPageConfigUtils'
import { useCreateOrUpdateCoursePhaseConfig } from './useCreateOrUpdateCoursePhaseConfig'

type AssessmentCardModel = Omit<SchemaConfigurationCardProps, 'schemas' | 'disabled' | 'isSaving'>

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

interface UseAssessmentSettingsCardStateResult {
  isSaving: boolean
  assessmentCard: AssessmentCardModel
  assessmentVisibility: AssessmentVisibilityModel
}

export const useAssessmentSettingsCardState = (): UseAssessmentSettingsCardStateResult => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [assessmentEnabled, setAssessmentEnabled] = useState<boolean>(true)
  const [assessmentSchemaId, setAssessmentSchemaId] = useState<string>('')
  const [start, setStart] = useState<Date | undefined>(undefined)
  const [deadline, setDeadline] = useState<Date | undefined>(undefined)
  const [evaluationResultsVisible, setEvaluationResultsVisible] = useState<boolean>(false)
  const [gradeSuggestionVisible, setGradeSuggestionVisible] = useState<boolean>(true)
  const [actionItemsVisible, setActionItemsVisible] = useState<boolean>(true)
  const [gradingSheetVisible, setGradingSheetVisible] = useState<boolean>(false)
  const { data: originalConfig } = useGetCoursePhaseConfig()

  useEffect(() => {
    if (originalConfig) {
      setAssessmentEnabled(originalConfig.assessmentEnabled ?? true)
      setAssessmentSchemaId(originalConfig.assessmentSchemaID || '')
      setStart(originalConfig.start ? new Date(originalConfig.start) : undefined)
      setDeadline(originalConfig.deadline ? new Date(originalConfig.deadline) : undefined)
      setEvaluationResultsVisible(originalConfig.evaluationResultsVisible || false)
      setGradeSuggestionVisible(originalConfig.gradeSuggestionVisible ?? true)
      setActionItemsVisible(originalConfig.actionItemsVisible ?? true)
      setGradingSheetVisible(originalConfig.gradingSheetVisible ?? false)
      return
    }

    setAssessmentEnabled(true)
    setAssessmentSchemaId('')
    setStart(undefined)
    setDeadline(undefined)
    setEvaluationResultsVisible(false)
    setGradeSuggestionVisible(true)
    setActionItemsVisible(true)
    setGradingSheetVisible(false)
  }, [originalConfig])

  const configMutation = useCreateOrUpdateCoursePhaseConfig({
    onSuccess: () => setError(undefined),
    onError: setError,
  })

  const { data: assessmentSchemaData } = useSchemaHasAssessmentData(assessmentSchemaId || undefined)
  const baseRequest = useMemo(() => buildRequestFromConfig(originalConfig), [originalConfig])

  const hasChanges = useMemo(
    () =>
      hasAssessmentCardChanges(
        {
          assessmentEnabled,
          assessmentSchemaId,
          start,
          deadline,
          evaluationResultsVisible,
          gradeSuggestionVisible,
          actionItemsVisible,
          gradingSheetVisible,
        },
        originalConfig,
      ),
    [
      actionItemsVisible,
      assessmentEnabled,
      assessmentSchemaId,
      deadline,
      evaluationResultsVisible,
      gradeSuggestionVisible,
      gradingSheetVisible,
      originalConfig,
      start,
    ],
  )

  const assessmentCard: AssessmentCardModel = {
    assessmentType: AssessmentType.ASSESSMENT,
    enabled: assessmentEnabled,
    onEnabledChange: setAssessmentEnabled,
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
    error,
    hasChanges,
    onSave: () => {
      if (!baseRequest) {
        setError('Settings are still loading. Please try again in a moment.')
        return
      }

      configMutation.mutate({
        ...baseRequest,
        assessmentEnabled,
        assessmentSchemaId,
        start,
        deadline,
        evaluationResultsVisible,
        gradeSuggestionVisible,
        actionItemsVisible,
        gradingSheetVisible,
      })
    },
    canSave: (!assessmentEnabled || Boolean(assessmentSchemaId)) && Boolean(baseRequest),
    onCreateSchemaError: setError,
  }

  return {
    isSaving: configMutation.isPending,
    assessmentCard,
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
