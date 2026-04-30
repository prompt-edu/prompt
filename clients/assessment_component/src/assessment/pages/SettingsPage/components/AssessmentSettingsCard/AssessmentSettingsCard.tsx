import { endOfDay, startOfDay } from 'date-fns'
import { Link } from 'react-router-dom'
import {
  Button,
  Card,
  DatePickerWithRange,
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@tumaet/prompt-ui-components'
import { CalendarRange, ExternalLink, FileStack, Lock } from 'lucide-react'

import { AssessmentType } from '../../../../interfaces/assessmentType'
import { useGetAllAssessmentSchemas } from '../../../hooks/useGetAllAssessmentSchemas'
import { schemaSectionContent } from '../../../schemaSectionContent'
import { useAssessmentSettingsCardState } from '../../hooks/useAssessmentSettingsCardState'
import { SettingsSwitchField } from '../SettingsSwitchField'
import { CreateAssessmentSchemaDialog } from '../CreateAssessmentSchemaDialog'
import { ErrorDisplay } from '../ErrorDisplay'
import { ReleaseResultsSection } from './components/ReleaseResultsSection'

export const AssessmentSettingsCard = () => {
  const { isSaving, assessmentCard, assessmentVisibility } = useAssessmentSettingsCardState()
  const {
    data: schemas,
    isPending: isSchemasPending,
    isError: isSchemasError,
  } = useGetAllAssessmentSchemas()

  const content = schemaSectionContent[AssessmentType.ASSESSMENT]
  const controlsDisabled = isSaving
  const schemaControlsDisabled =
    controlsDisabled ||
    isSchemasPending ||
    isSchemasError ||
    (assessmentCard.hasAssessmentData ?? false)
  const timeframeControlsDisabled = controlsDisabled || isSchemasPending || isSchemasError
  const saveDisabled = controlsDisabled || !assessmentCard.hasChanges || !assessmentCard.canSave

  return (
    <Card className='border-border shadow-xs'>
      <div className='space-y-6 p-6'>
        <div className='space-y-2'>
          <div className='flex items-center gap-2'>
            <h2 className='text-xl font-semibold text-foreground'>{content.title}</h2>
            {assessmentCard.hasAssessmentData && (
              <span
                className={[
                  'inline-flex items-center gap-1 rounded-full bg-amber-100 px-3 py-1',
                  'text-xs font-medium text-amber-800 dark:bg-amber-900/30',
                  'dark:text-amber-300',
                ].join(' ')}
              >
                <Lock className='h-3.5 w-3.5' />
                Locked by submitted data
              </span>
            )}
          </div>
          <p className='max-w-3xl text-sm leading-6 text-muted-foreground'>{content.summary}</p>
        </div>

        <ErrorDisplay error={assessmentCard.error} />

        <div className='grid gap-6 xl:grid-cols-[1.2fr_0.8fr]'>
          <div className='space-y-3'>
            <div className='flex items-center gap-2 text-sm font-semibold text-foreground'>
              <FileStack className='h-4 w-4' />
              <span>{content.schemaLabel}</span>
            </div>
            <p className='text-sm leading-6 text-muted-foreground'>{content.schemaHint}</p>
            {isSchemasError && (
              <p className='text-xs text-destructive'>
                Assessment schemas could not be loaded. Please refresh and try again.
              </p>
            )}
            <div className='flex flex-col gap-3 lg:flex-row lg:items-center'>
              <div className='flex flex-1 gap-2'>
                <Select
                  value={assessmentCard.schemaId}
                  onValueChange={assessmentCard.onSchemaIdChange}
                  disabled={schemaControlsDisabled}
                >
                  <SelectTrigger className='flex-1'>
                    <SelectValue placeholder='Select a schema...' />
                  </SelectTrigger>
                  <SelectContent>
                    {(schemas ?? []).map((schema) => (
                      <SelectItem key={schema.id} value={schema.id}>
                        {schema.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <CreateAssessmentSchemaDialog
                  onError={assessmentCard.onCreateSchemaError}
                  disabled={schemaControlsDisabled}
                />
              </div>

              {assessmentCard.schemaId && assessmentCard.canOpenDetails ? (
                <Button asChild variant='secondary' className='lg:self-stretch'>
                  <Link to={assessmentCard.detailPath}>
                    Open schema details
                    <ExternalLink className='ml-2 h-4 w-4' />
                  </Link>
                </Button>
              ) : (
                <Button variant='secondary' disabled className='lg:self-stretch'>
                  {assessmentCard.detailButtonLabel}
                  <ExternalLink className='ml-2 h-4 w-4' />
                </Button>
              )}
            </div>
          </div>

          <div className='space-y-3'>
            <div className='flex items-center gap-2 text-sm font-semibold text-foreground'>
              <CalendarRange className='h-4 w-4' />
              <span>{content.timeframeLabel}</span>
            </div>
            <p className='text-sm leading-6 text-muted-foreground'>{content.timeframeHint}</p>
            <div className={`${timeframeControlsDisabled ? 'pointer-events-none opacity-60' : ''}`}>
              <DatePickerWithRange
                date={{
                  from: assessmentCard.startDate,
                  to: assessmentCard.deadline,
                }}
                setDate={(dateRange) => {
                  assessmentCard.onStartDateChange(
                    dateRange?.from ? startOfDay(dateRange.from) : undefined,
                  )
                  assessmentCard.onDeadlineChange(
                    dateRange?.to ? endOfDay(dateRange.to) : undefined,
                  )
                }}
              />
            </div>
          </div>
        </div>

        <div className='grid gap-6 xl:grid-cols-2'>
          <div className='space-y-4'>
            <div className='space-y-1'>
              <h3 className='text-sm font-semibold text-foreground'>
                Student visibility after release
              </h3>
            </div>

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
            <div className='space-y-1'>
              <h3 className='text-sm font-semibold text-foreground'>
                Assessment workflow visibility
              </h3>
            </div>

            <SettingsSwitchField
              checked={assessmentVisibility.evaluationResultsVisible}
              onCheckedChange={assessmentVisibility.setEvaluationResultsVisible}
              disabled={isSaving}
              title='Show evaluation results before submission'
              description='Assessment authors can review self-, peer-, and student-to-tutor evaluation results before they finalize the assessment.'
            />

            <ReleaseResultsSection isSaving={isSaving} />
          </div>
        </div>

        <div className='flex flex-wrap items-center gap-3 border-t border-border pt-4'>
          <p className='flex-1 text-xs leading-5 text-muted-foreground'>
            Save this card to apply its configuration changes for the current course phase.
          </p>
          <Button
            onClick={assessmentCard.onSave}
            disabled={saveDisabled}
            className='ml-auto min-w-[160px]'
          >
            {isSaving ? 'Saving...' : 'Save Assessment'}
          </Button>
        </div>
      </div>
    </Card>
  )
}
