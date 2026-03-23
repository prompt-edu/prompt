import { useState } from 'react'
import { Button, ErrorPage, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'

import { AssessmentType } from '../../interfaces/assessmentType'
import {
  CoursePhaseConfig,
  CreateOrUpdateCoursePhaseConfigRequest,
} from '../../interfaces/coursePhaseConfig'
import { useParticipationStore } from '../../zustand/useParticipationStore'
import { useCoursePhaseConfigStore } from '../../zustand/useCoursePhaseConfigStore'
import { useGetAllAssessmentCompletions } from '../hooks/useGetAllAssessmentCompletions'
import { useSchemaHasAssessmentData } from './hooks/usePhaseHasAssessmentData'
import { useReleaseResults } from './hooks/useReleaseResults'
import { ReleaseConfirmationDialog } from './components/ReleaseConfirmationDialog'
import { SchemaConfigurationCard } from './components/SchemaConfigurationCard'
import { SettingsSwitchField } from './components/SettingsSwitchField'
import { useGetAllAssessmentSchemas } from './components/CoursePhaseConfigSelection/hooks/useGetAllAssessmentSchemas'
import { useCreateOrUpdateCoursePhaseConfig } from './components/CoursePhaseConfigSelection/hooks/useCreateOrUpdateCoursePhaseConfig'
import { useCoursePhaseConfigForm } from './components/CoursePhaseConfigSelection/hooks/useCoursePhaseConfigForm'
import { useEvaluationOptions } from './components/CoursePhaseConfigSelection/hooks/useEvaluationOptions'

const toDate = (value?: Date): Date | undefined => {
  if (!value) return undefined

  return new Date(value)
}

const areDatesEqual = (left?: Date, right?: Date) => left?.getTime() === right?.getTime()

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

export const SettingsPage = () => {
  const [error, setError] = useState<string | undefined>(undefined)
  const [activeErrorCard, setActiveErrorCard] = useState<AssessmentType | undefined>(undefined)
  const [showReleaseDialog, setShowReleaseDialog] = useState(false)

  const { participations } = useParticipationStore()
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
  } = useEvaluationOptions()

  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()

  const { data: assessmentCompletions } = useGetAllAssessmentCompletions()
  const { mutate: releaseResults, isPending: isReleasing } = useReleaseResults()

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

  if (isSchemasError) {
    return <ErrorPage />
  }

  if (isSchemasPending) {
    return (
      <div className='flex h-64 items-center justify-center'>
        <Loader2 className='h-12 w-12 animate-spin text-primary' />
      </div>
    )
  }

  const totalAssessments = participations.length
  const completedAssessments =
    assessmentCompletions?.filter((completion) => completion.completed).length ?? 0
  const allAssessmentsCompleted = totalAssessments > 0 && completedAssessments === totalAssessments

  const currentRequest: CreateOrUpdateCoursePhaseConfigRequest = {
    assessmentSchemaId,
    start,
    deadline,
    selfEvaluationEnabled,
    selfEvaluationSchema: selfEvaluationSchema || undefined,
    selfEvaluationStart,
    selfEvaluationDeadline,
    peerEvaluationEnabled,
    peerEvaluationSchema: peerEvaluationSchema || undefined,
    peerEvaluationStart,
    peerEvaluationDeadline,
    tutorEvaluationEnabled,
    tutorEvaluationSchema: tutorEvaluationSchema || undefined,
    tutorEvaluationStart,
    tutorEvaluationDeadline,
    evaluationResultsVisible,
    gradeSuggestionVisible,
    actionItemsVisible,
    gradingSheetVisible,
  }

  const buildBaseRequest = () =>
    originalConfig ? buildRequestFromConfig(originalConfig) : currentRequest

  const handleCardError = (assessmentType: AssessmentType, nextError: string | undefined) => {
    setActiveErrorCard(assessmentType)
    setError(nextError)
  }

  const handleSave = (
    assessmentType: AssessmentType,
    request: CreateOrUpdateCoursePhaseConfigRequest,
  ) => {
    setActiveErrorCard(assessmentType)
    configMutation.mutate(request)
  }

  const confirmRelease = () => {
    releaseResults(undefined, {
      onSuccess: () => setShowReleaseDialog(false),
    })
  }

  const assessmentHasChanges = hasMainConfigChanges(originalConfig)
  const selfHasChanges = !originalConfig
    ? Boolean(
        selfEvaluationEnabled ||
        selfEvaluationSchema ||
        selfEvaluationStart ||
        selfEvaluationDeadline,
      )
    : selfEvaluationEnabled !== (originalConfig.selfEvaluationEnabled ?? false) ||
      selfEvaluationSchema !== (originalConfig.selfEvaluationSchema || '') ||
      !areDatesEqual(selfEvaluationStart, toDate(originalConfig.selfEvaluationStart)) ||
      !areDatesEqual(selfEvaluationDeadline, toDate(originalConfig.selfEvaluationDeadline))
  const peerHasChanges = !originalConfig
    ? Boolean(
        peerEvaluationEnabled ||
        peerEvaluationSchema ||
        peerEvaluationStart ||
        peerEvaluationDeadline,
      )
    : peerEvaluationEnabled !== (originalConfig.peerEvaluationEnabled ?? false) ||
      peerEvaluationSchema !== (originalConfig.peerEvaluationSchema || '') ||
      !areDatesEqual(peerEvaluationStart, toDate(originalConfig.peerEvaluationStart)) ||
      !areDatesEqual(peerEvaluationDeadline, toDate(originalConfig.peerEvaluationDeadline))
  const tutorHasChanges = !originalConfig
    ? Boolean(
        tutorEvaluationEnabled ||
        tutorEvaluationSchema ||
        tutorEvaluationStart ||
        tutorEvaluationDeadline,
      )
    : tutorEvaluationEnabled !== (originalConfig.tutorEvaluationEnabled ?? false) ||
      tutorEvaluationSchema !== (originalConfig.tutorEvaluationSchema || '') ||
      !areDatesEqual(tutorEvaluationStart, toDate(originalConfig.tutorEvaluationStart)) ||
      !areDatesEqual(tutorEvaluationDeadline, toDate(originalConfig.tutorEvaluationDeadline))

  const assessmentDetailReady =
    assessmentSchemaId === (originalConfig?.assessmentSchemaID ?? '') &&
    Boolean(originalConfig?.assessmentSchemaID)
  const selfDetailReady =
    selfEvaluationEnabled &&
    selfEvaluationSchema === (originalConfig?.selfEvaluationSchema ?? '') &&
    Boolean(originalConfig?.selfEvaluationEnabled) &&
    Boolean(originalConfig?.selfEvaluationSchema)
  const peerDetailReady =
    peerEvaluationEnabled &&
    peerEvaluationSchema === (originalConfig?.peerEvaluationSchema ?? '') &&
    Boolean(originalConfig?.peerEvaluationEnabled) &&
    Boolean(originalConfig?.peerEvaluationSchema)
  const tutorDetailReady =
    tutorEvaluationEnabled &&
    tutorEvaluationSchema === (originalConfig?.tutorEvaluationSchema ?? '') &&
    Boolean(originalConfig?.tutorEvaluationEnabled) &&
    Boolean(originalConfig?.tutorEvaluationSchema)

  return (
    <div className='space-y-6'>
      <ManagementPageHeader>Assessment Settings</ManagementPageHeader>

      <SchemaConfigurationCard
        assessmentType={AssessmentType.ASSESSMENT}
        enabled
        schemaId={assessmentSchemaId}
        onSchemaIdChange={setAssessmentSchemaId}
        startDate={start}
        onStartDateChange={setStart}
        deadline={deadline}
        onDeadlineChange={setDeadline}
        schemas={schemas ?? []}
        detailPath={`schema/${AssessmentType.ASSESSMENT}`}
        canOpenDetails={assessmentDetailReady}
        detailButtonLabel={
          assessmentSchemaId ? 'Save this card to open schema details' : 'Select a schema first'
        }
        hasAssessmentData={assessmentSchemaData?.hasAssessmentData ?? false}
        disabled={configMutation.isPending}
        error={activeErrorCard === AssessmentType.ASSESSMENT ? error : undefined}
        hasChanges={assessmentHasChanges}
        isSaving={configMutation.isPending}
        onSave={() =>
          handleSave(AssessmentType.ASSESSMENT, {
            ...buildBaseRequest(),
            assessmentSchemaId,
            start,
            deadline,
            evaluationResultsVisible,
            gradeSuggestionVisible,
            actionItemsVisible,
            gradingSheetVisible,
          })
        }
        canSave={Boolean(assessmentSchemaId)}
        onCreateSchemaError={(nextError) => handleCardError(AssessmentType.ASSESSMENT, nextError)}
        showToggle={false}
      >
        <div className='grid gap-6 xl:grid-cols-2'>
          <div className='space-y-4'>
            <div className='space-y-1'>
              <h3 className='text-sm font-semibold text-slate-900'>
                Student visibility after release
              </h3>
              <p className='text-sm leading-6 text-slate-600'>
                Choose which parts of the final assessment become visible to students after you
                release results for this phase.
              </p>
            </div>

            <SettingsSwitchField
              checked={gradingSheetVisible}
              onCheckedChange={setGradingSheetVisible}
              disabled={configMutation.isPending}
              title='Show assessment sheet'
              description='Students can inspect the grading sheet, including score levels, examples, and assessor comments.'
            />
            <SettingsSwitchField
              checked={gradeSuggestionVisible}
              onCheckedChange={setGradeSuggestionVisible}
              disabled={configMutation.isPending}
              title='Show grade suggestions'
              description='Students can see the proposed grade and the final written feedback attached to their assessment.'
            />
            <SettingsSwitchField
              checked={actionItemsVisible}
              onCheckedChange={setActionItemsVisible}
              disabled={configMutation.isPending}
              title='Show action items'
              description='Students can see the follow-up actions or recommendations assessors recorded for them.'
            />
          </div>

          <div className='space-y-4'>
            <div className='space-y-1'>
              <h3 className='text-sm font-semibold text-slate-900'>Assessor workflow visibility</h3>
              <p className='text-sm leading-6 text-slate-600'>
                Control whether assessors can inspect related evaluation results before they submit
                the final assessment.
              </p>
            </div>

            <SettingsSwitchField
              checked={evaluationResultsVisible}
              onCheckedChange={setEvaluationResultsVisible}
              disabled={configMutation.isPending}
              title='Show evaluation results before submission'
              description='Assessors can review self-, peer-, and tutor-evaluation results before they finalize the assessment.'
            />

            <div className='rounded-xl border border-slate-200 bg-slate-50/60 p-4'>
              {originalConfig?.resultsReleased ? (
                <div className='space-y-1'>
                  <h3 className='text-sm font-semibold text-emerald-900'>Results released</h3>
                  <p className='text-sm leading-6 text-emerald-700'>
                    Students in this phase can already view the released assessment results based on
                    the visibility settings above.
                  </p>
                </div>
              ) : (
                <div className='space-y-4'>
                  <div className='space-y-1'>
                    <h3 className='text-sm font-semibold text-slate-900'>Release results</h3>
                    <p className='text-sm leading-6 text-slate-600'>
                      Release final assessment results to students after every assessment is marked
                      as final. This action cannot be undone.
                    </p>
                  </div>

                  <Button
                    onClick={() => setShowReleaseDialog(true)}
                    disabled={configMutation.isPending || isReleasing || !allAssessmentsCompleted}
                    className='w-full'
                  >
                    {isReleasing
                      ? 'Releasing...'
                      : `Release Results (${completedAssessments}/${totalAssessments} final)`}
                  </Button>

                  {!allAssessmentsCompleted && (
                    <p className='text-xs leading-5 text-slate-500'>
                      All assessments must be marked as final before results can be released.
                    </p>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      </SchemaConfigurationCard>

      <SchemaConfigurationCard
        assessmentType={AssessmentType.SELF}
        enabled={selfEvaluationEnabled}
        onEnabledChange={setSelfEvaluationEnabled}
        schemaId={selfEvaluationSchema}
        onSchemaIdChange={setSelfEvaluationSchema}
        startDate={selfEvaluationStart}
        onStartDateChange={setSelfEvaluationStart}
        deadline={selfEvaluationDeadline}
        onDeadlineChange={setSelfEvaluationDeadline}
        schemas={schemas ?? []}
        detailPath={`schema/${AssessmentType.SELF}`}
        canOpenDetails={selfDetailReady}
        detailButtonLabel={
          selfEvaluationSchema ? 'Save this card to open schema details' : 'Select a schema first'
        }
        hasAssessmentData={selfSchemaData?.hasAssessmentData ?? false}
        disabled={configMutation.isPending}
        error={activeErrorCard === AssessmentType.SELF ? error : undefined}
        hasChanges={selfHasChanges}
        isSaving={configMutation.isPending}
        onSave={() =>
          handleSave(AssessmentType.SELF, {
            ...buildBaseRequest(),
            selfEvaluationEnabled,
            selfEvaluationSchema: selfEvaluationSchema || undefined,
            selfEvaluationStart,
            selfEvaluationDeadline,
          })
        }
        canSave={!selfEvaluationEnabled || Boolean(selfEvaluationSchema)}
        onCreateSchemaError={(nextError) => handleCardError(AssessmentType.SELF, nextError)}
      />

      <SchemaConfigurationCard
        assessmentType={AssessmentType.PEER}
        enabled={peerEvaluationEnabled}
        onEnabledChange={setPeerEvaluationEnabled}
        schemaId={peerEvaluationSchema}
        onSchemaIdChange={setPeerEvaluationSchema}
        startDate={peerEvaluationStart}
        onStartDateChange={setPeerEvaluationStart}
        deadline={peerEvaluationDeadline}
        onDeadlineChange={setPeerEvaluationDeadline}
        schemas={schemas ?? []}
        detailPath={`schema/${AssessmentType.PEER}`}
        canOpenDetails={peerDetailReady}
        detailButtonLabel={
          peerEvaluationSchema ? 'Save this card to open schema details' : 'Select a schema first'
        }
        hasAssessmentData={peerSchemaData?.hasAssessmentData ?? false}
        disabled={configMutation.isPending}
        error={activeErrorCard === AssessmentType.PEER ? error : undefined}
        hasChanges={peerHasChanges}
        isSaving={configMutation.isPending}
        onSave={() =>
          handleSave(AssessmentType.PEER, {
            ...buildBaseRequest(),
            peerEvaluationEnabled,
            peerEvaluationSchema: peerEvaluationSchema || undefined,
            peerEvaluationStart,
            peerEvaluationDeadline,
          })
        }
        canSave={!peerEvaluationEnabled || Boolean(peerEvaluationSchema)}
        onCreateSchemaError={(nextError) => handleCardError(AssessmentType.PEER, nextError)}
      />

      <SchemaConfigurationCard
        assessmentType={AssessmentType.TUTOR}
        enabled={tutorEvaluationEnabled}
        onEnabledChange={setTutorEvaluationEnabled}
        schemaId={tutorEvaluationSchema}
        onSchemaIdChange={setTutorEvaluationSchema}
        startDate={tutorEvaluationStart}
        onStartDateChange={setTutorEvaluationStart}
        deadline={tutorEvaluationDeadline}
        onDeadlineChange={setTutorEvaluationDeadline}
        schemas={schemas ?? []}
        detailPath={`schema/${AssessmentType.TUTOR}`}
        canOpenDetails={tutorDetailReady}
        detailButtonLabel={
          tutorEvaluationSchema ? 'Save this card to open schema details' : 'Select a schema first'
        }
        hasAssessmentData={tutorSchemaData?.hasAssessmentData ?? false}
        disabled={configMutation.isPending}
        error={activeErrorCard === AssessmentType.TUTOR ? error : undefined}
        hasChanges={tutorHasChanges}
        isSaving={configMutation.isPending}
        onSave={() =>
          handleSave(AssessmentType.TUTOR, {
            ...buildBaseRequest(),
            tutorEvaluationEnabled,
            tutorEvaluationSchema: tutorEvaluationSchema || undefined,
            tutorEvaluationStart,
            tutorEvaluationDeadline,
          })
        }
        canSave={!tutorEvaluationEnabled || Boolean(tutorEvaluationSchema)}
        onCreateSchemaError={(nextError) => handleCardError(AssessmentType.TUTOR, nextError)}
      />

      <ReleaseConfirmationDialog
        open={showReleaseDialog}
        onOpenChange={setShowReleaseDialog}
        onConfirm={confirmRelease}
        isReleasing={isReleasing}
        completedAssessments={completedAssessments}
        totalAssessments={totalAssessments}
      />
    </div>
  )
}
