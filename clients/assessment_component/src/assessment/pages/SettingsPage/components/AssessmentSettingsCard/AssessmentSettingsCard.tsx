import { useGetAllAssessmentSchemas } from '../../../hooks/useGetAllAssessmentSchemas'
import { useAssessmentSettingsCardState } from '../../hooks/useAssessmentSettingsCardState'
import { SchemaConfigurationCard } from '../SchemaConfigurationCard'
import { SettingsSwitchField } from '../SettingsSwitchField'

export const AssessmentSettingsCard = () => {
  const { isSaving, assessmentCard, assessmentVisibility } = useAssessmentSettingsCardState()
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
          <h3 className='text-sm font-semibold text-foreground'>Assessment workflow visibility</h3>

          <SettingsSwitchField
            checked={assessmentVisibility.evaluationResultsVisible}
            onCheckedChange={assessmentVisibility.setEvaluationResultsVisible}
            disabled={isSaving}
            title='Show evaluation results before submission'
            description='Assessment authors can review self-, peer-, and student-to-tutor evaluation results before they finalize the assessment.'
          />
        </div>
      </div>
    </SchemaConfigurationCard>
  )
}
