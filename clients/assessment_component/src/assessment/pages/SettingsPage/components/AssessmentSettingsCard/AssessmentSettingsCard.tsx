import { useGetAllAssessmentSchemas } from '../../../hooks/useGetAllAssessmentSchemas'
import { useTutorLabel } from '../../../hooks/useTutorLabel'
import { useAssessmentSettingsCardState } from '../../hooks/useAssessmentSettingsCardState'
import { SchemaConfigurationCard } from '../SchemaConfigurationCard'
import { SettingsSwitchField } from '../SettingsSwitchField'

export const AssessmentSettingsCard = () => {
  const { isSaving, assessmentCard, assessmentVisibility } = useAssessmentSettingsCardState()
  const tutorLabel = useTutorLabel()
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()

  return (
    <SchemaConfigurationCard
      {...assessmentCard}
      schemas={schemas ?? []}
      disabled={isSchemasPending || isSchemasError}
      isSaving={isSaving}
    >
      {isSchemasError && (
        <p className='text-xs text-destructive'>
          Assessment schemas could not be loaded. Please refresh and try again.
        </p>
      )}

      <div className='grid gap-6 xl:grid-cols-2'>
        <div className='space-y-4'>
          <h3 className='text-sm font-semibold text-foreground'>
            Student visibility after release
          </h3>

          <SettingsSwitchField
            checked={assessmentVisibility.gradingSheetVisible}
            onCheckedChange={assessmentVisibility.setGradingSheetVisible}
            disabled={isSaving}
            title='Show assessment sheet'
            description='Students can inspect the grading sheet, including score levels, examples, and comments.'
          />
          <SettingsSwitchField
            checked={assessmentVisibility.gradeSuggestionVisible}
            onCheckedChange={assessmentVisibility.setGradeSuggestionVisible}
            disabled={isSaving}
            title='Show grade suggestions'
            description='Students can see the proposed grade and the final written feedback attached to their assessment.'
          />
          <SettingsSwitchField
            checked={assessmentVisibility.actionItemsVisible}
            onCheckedChange={assessmentVisibility.setActionItemsVisible}
            disabled={isSaving}
            title='Show action items'
            description='Students can see the action-items recorded for them.'
          />
        </div>

        <div className='space-y-4'>
          <h3 className='text-sm font-semibold text-foreground'>Assessment workflow</h3>

          <SettingsSwitchField
            checked={assessmentVisibility.independentAssessmentEnabled}
            onCheckedChange={assessmentVisibility.setIndependentAssessmentEnabled}
            disabled={isSaving}
            title='Independent assessments'
            description='Every assessor first scores students on their own in a "My assessment" tab. The final assessment then shows all scores side by side so they can be merged.'
          />
          <SettingsSwitchField
            checked={assessmentVisibility.evaluationResultsVisible}
            onCheckedChange={assessmentVisibility.setEvaluationResultsVisible}
            disabled={isSaving}
            title='Show evaluation results before submission'
            description={`Assessment authors can review self-, peer-, and student-to-${tutorLabel.text} evaluation results, and other assessors' independent scores, before they finalize the assessment.`}
          />
        </div>
      </div>
    </SchemaConfigurationCard>
  )
}
