import { Button } from '@tumaet/prompt-ui-components'

import type { AssessmentSchema } from '../../../../interfaces/assessmentSchema'
import type { SchemaConfigurationCardProps } from '../SchemaConfigurationCard'
import { SchemaConfigurationCard } from '../SchemaConfigurationCard'
import { SettingsSwitchField } from '../SettingsSwitchField'
import { ReleaseConfirmationDialog } from '../ReleaseConfirmationDialog'

type SchemaCardConfig = Omit<SchemaConfigurationCardProps, 'schemas' | 'disabled' | 'isSaving'>

interface CoursePhaseConfigSelectionProps {
  schemas: AssessmentSchema[]
  disabled?: boolean
  isSaving: boolean
  assessmentCard: SchemaCardConfig
  selfCard: SchemaCardConfig
  peerCard: SchemaCardConfig
  tutorCard: SchemaCardConfig
  gradingSheetVisible: boolean
  onGradingSheetVisibleChange: (checked: boolean) => void
  gradeSuggestionVisible: boolean
  onGradeSuggestionVisibleChange: (checked: boolean) => void
  actionItemsVisible: boolean
  onActionItemsVisibleChange: (checked: boolean) => void
  evaluationResultsVisible: boolean
  onEvaluationResultsVisibleChange: (checked: boolean) => void
  resultsReleased: boolean
  showReleaseDialog: boolean
  onShowReleaseDialogChange: (open: boolean) => void
  onConfirmRelease: () => void
  isReleasing: boolean
  completedAssessments: number
  totalAssessments: number
  allAssessmentsCompleted: boolean
}

export const CoursePhaseConfigSelection = ({
  schemas,
  disabled = false,
  isSaving,
  assessmentCard,
  selfCard,
  peerCard,
  tutorCard,
  gradingSheetVisible,
  onGradingSheetVisibleChange,
  gradeSuggestionVisible,
  onGradeSuggestionVisibleChange,
  actionItemsVisible,
  onActionItemsVisibleChange,
  evaluationResultsVisible,
  onEvaluationResultsVisibleChange,
  resultsReleased,
  showReleaseDialog,
  onShowReleaseDialogChange,
  onConfirmRelease,
  isReleasing,
  completedAssessments,
  totalAssessments,
  allAssessmentsCompleted,
}: CoursePhaseConfigSelectionProps) => {
  const controlsDisabled = disabled || isSaving

  return (
    <div className='space-y-6'>
      <SchemaConfigurationCard
        {...assessmentCard}
        schemas={schemas}
        disabled={controlsDisabled}
        isSaving={isSaving}
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
              onCheckedChange={onGradingSheetVisibleChange}
              disabled={controlsDisabled}
              title='Show assessment sheet'
              description='Students can inspect the grading sheet, including score levels, examples, and assessor comments.'
            />
            <SettingsSwitchField
              checked={gradeSuggestionVisible}
              onCheckedChange={onGradeSuggestionVisibleChange}
              disabled={controlsDisabled}
              title='Show grade suggestions'
              description='Students can see the proposed grade and the final written feedback attached to their assessment.'
            />
            <SettingsSwitchField
              checked={actionItemsVisible}
              onCheckedChange={onActionItemsVisibleChange}
              disabled={controlsDisabled}
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
              onCheckedChange={onEvaluationResultsVisibleChange}
              disabled={controlsDisabled}
              title='Show evaluation results before submission'
              description='Assessors can review self-, peer-, and tutor-evaluation results before they finalize the assessment.'
            />

            <div className='rounded-xl border border-slate-200 bg-slate-50/60 p-4'>
              {resultsReleased ? (
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
                    onClick={() => onShowReleaseDialogChange(true)}
                    disabled={controlsDisabled || isReleasing || !allAssessmentsCompleted}
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
        {...selfCard}
        schemas={schemas}
        disabled={controlsDisabled}
        isSaving={isSaving}
      />

      <SchemaConfigurationCard
        {...peerCard}
        schemas={schemas}
        disabled={controlsDisabled}
        isSaving={isSaving}
      />

      <SchemaConfigurationCard
        {...tutorCard}
        schemas={schemas}
        disabled={controlsDisabled}
        isSaving={isSaving}
      />

      <ReleaseConfirmationDialog
        open={showReleaseDialog}
        onOpenChange={onShowReleaseDialogChange}
        onConfirm={onConfirmRelease}
        isReleasing={isReleasing}
        completedAssessments={completedAssessments}
        totalAssessments={totalAssessments}
      />
    </div>
  )
}

export type { SchemaCardConfig }
