import { useEffect, useMemo, useState } from 'react'

import { useCoursePhaseConfigStore } from '../../../zustand/useCoursePhaseConfigStore'
import { useSchemaHasAssessmentData } from '../../hooks/useSchemaHasAssessmentData'
import type { SchemaConfigurationCardProps } from '../components/SchemaConfigurationCard'
import {
  buildEvaluationPatch,
  buildRequestFromConfig,
  createDetailButtonLabel,
  EvaluationAssessmentType,
  getEvaluationOriginalSnapshot,
  hasEvaluationCardChanges,
  isEvaluationDetailReady,
} from './settingsPageConfigUtils'
import { useCreateOrUpdateCoursePhaseConfig } from './useCreateOrUpdateCoursePhaseConfig'

type EvaluationCardModel = Omit<SchemaConfigurationCardProps, 'schemas' | 'disabled' | 'isSaving'>

interface UseEvaluationSettingsCardStateResult {
  isSaving: boolean
  card: EvaluationCardModel
}

export const useEvaluationSettingsCardState = (
  assessmentType: EvaluationAssessmentType,
): UseEvaluationSettingsCardStateResult => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [enabled, setEnabled] = useState<boolean>(false)
  const [schema, setSchema] = useState<string>('')
  const [start, setStart] = useState<Date | undefined>(undefined)
  const [deadline, setDeadline] = useState<Date | undefined>(undefined)
  const { coursePhaseConfig: originalConfig } = useCoursePhaseConfigStore()

  const originalSnapshot = useMemo(
    () => getEvaluationOriginalSnapshot(originalConfig, assessmentType),
    [assessmentType, originalConfig],
  )

  useEffect(() => {
    setEnabled(originalSnapshot.enabled ?? false)
    setSchema(originalSnapshot.schema || '')
    setStart(originalSnapshot.start ? new Date(originalSnapshot.start) : undefined)
    setDeadline(originalSnapshot.deadline ? new Date(originalSnapshot.deadline) : undefined)
  }, [
    originalSnapshot.deadline,
    originalSnapshot.enabled,
    originalSnapshot.schema,
    originalSnapshot.start,
  ])

  const configMutation = useCreateOrUpdateCoursePhaseConfig({
    onSuccess: () => setError(undefined),
    onError: setError,
  })

  const { data: schemaHasAssessmentData } = useSchemaHasAssessmentData(
    enabled ? schema || undefined : undefined,
  )
  const baseRequest = useMemo(() => buildRequestFromConfig(originalConfig), [originalConfig])

  const hasChanges = useMemo(
    () =>
      hasEvaluationCardChanges(
        {
          enabled,
          schema,
          start,
          deadline,
        },
        originalSnapshot,
      ),
    [deadline, enabled, originalSnapshot, schema, start],
  )

  const card: EvaluationCardModel = {
    assessmentType,
    enabled,
    onEnabledChange: setEnabled,
    schemaId: schema,
    onSchemaIdChange: setSchema,
    startDate: start,
    onStartDateChange: setStart,
    deadline,
    onDeadlineChange: setDeadline,
    detailPath: schema ? `schema/${schema}` : '',
    canOpenDetails: isEvaluationDetailReady(
      enabled,
      schema,
      originalSnapshot.enabled,
      originalSnapshot.schema,
    ),
    detailButtonLabel: createDetailButtonLabel(schema),
    hasAssessmentData: schemaHasAssessmentData?.hasAssessmentData ?? false,
    error,
    hasChanges,
    onSave: () => {
      if (!baseRequest) {
        setError('Settings are still loading. Please try again in a moment.')
        return
      }

      configMutation.mutate({
        ...baseRequest,
        ...buildEvaluationPatch(assessmentType, {
          enabled,
          schema,
          start,
          deadline,
        }),
      })
    },
    canSave: (!enabled || Boolean(schema)) && Boolean(baseRequest),
    onCreateSchemaError: setError,
  }

  return {
    isSaving: configMutation.isPending,
    card,
  }
}
