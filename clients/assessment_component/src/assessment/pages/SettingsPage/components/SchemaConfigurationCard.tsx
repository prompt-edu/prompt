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

import { AssessmentType } from '../../../interfaces/assessmentType'
import { AssessmentSchema } from '../../../interfaces/assessmentSchema'
import { schemaSectionContent } from '../../schemaSectionContent'
import { SettingsSwitchField } from './SettingsSwitchField'
import { CreateAssessmentSchemaDialog } from './CreateAssessmentSchemaDialog'
import { ErrorDisplay } from './ErrorDisplay'

export interface SchemaConfigurationCardProps {
  assessmentType: AssessmentType
  enabled: boolean
  onEnabledChange?: (enabled: boolean) => void
  schemaId: string
  onSchemaIdChange: (schemaId: string) => void
  startDate?: Date
  onStartDateChange: (date: Date | undefined) => void
  deadline?: Date
  onDeadlineChange: (date: Date | undefined) => void
  schemas: AssessmentSchema[]
  detailPath: string
  canOpenDetails?: boolean
  detailButtonLabel?: string
  hasAssessmentData?: boolean
  disabled?: boolean
  error?: string
  hasChanges: boolean
  isSaving: boolean
  onSave: () => void
  canSave: boolean
  onCreateSchemaError: (error: string | undefined) => void
  renderAsCard?: boolean
  showToggle?: boolean
  children?: React.ReactNode
}

export const SchemaConfigurationCard = ({
  assessmentType,
  enabled,
  onEnabledChange,
  schemaId,
  onSchemaIdChange,
  startDate,
  onStartDateChange,
  deadline,
  onDeadlineChange,
  schemas,
  detailPath,
  canOpenDetails = false,
  detailButtonLabel = 'Save this card to open schema details',
  hasAssessmentData = false,
  disabled = false,
  error,
  hasChanges,
  isSaving,
  onSave,
  canSave,
  onCreateSchemaError,
  renderAsCard = true,
  showToggle = true,
  children,
}: SchemaConfigurationCardProps) => {
  const content = schemaSectionContent[assessmentType]
  const isRequired = assessmentType === AssessmentType.ASSESSMENT
  const isActive = isRequired || enabled
  const controlsDisabled = disabled || isSaving
  const enableSwitchDisabled = isRequired || controlsDisabled || hasAssessmentData
  const schemaControlsDisabled = controlsDisabled || hasAssessmentData || !isActive
  const timeframeControlsDisabled = controlsDisabled || !isActive
  const saveDisabled = controlsDisabled || !hasChanges || !canSave

  const contentBlock = (
    <div className={renderAsCard ? 'space-y-6 p-6' : 'space-y-6'}>
      <div className='flex flex-wrap items-start gap-4'>
        <div className='min-w-0 flex-1 space-y-2'>
          <div className='flex items-center gap-2'>
            <h2 className='text-xl font-semibold text-foreground'>{content.title}</h2>
            {hasAssessmentData && (
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

        {showToggle && (
          <div className='ml-auto w-fit max-w-sm'>
            <SettingsSwitchField
              checked={enabled}
              onCheckedChange={onEnabledChange}
              title={content.toggleLabel}
              description={content.toggleHint}
              disabled={enableSwitchDisabled}
            />
          </div>
        )}
      </div>

      <ErrorDisplay error={error} />

      {isActive && (
        <div className='grid gap-6 xl:grid-cols-[1.2fr_0.8fr]'>
          <div className='space-y-3'>
            <div className='flex items-center gap-2 text-sm font-semibold text-foreground'>
              <FileStack className='h-4 w-4' />
              <span>{content.schemaLabel}</span>
            </div>
            <p className='text-sm leading-6 text-muted-foreground'>{content.schemaHint}</p>
            <div className='flex flex-col gap-3 lg:flex-row lg:items-center'>
              <div className='flex flex-1 gap-2'>
                <Select
                  value={schemaId}
                  onValueChange={onSchemaIdChange}
                  disabled={schemaControlsDisabled}
                >
                  <SelectTrigger className='flex-1'>
                    <SelectValue placeholder='Select a schema...' />
                  </SelectTrigger>
                  <SelectContent>
                    {schemas.map((schema) => (
                      <SelectItem key={schema.id} value={schema.id}>
                        {schema.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <CreateAssessmentSchemaDialog
                  onError={onCreateSchemaError}
                  disabled={schemaControlsDisabled}
                />
              </div>

              {schemaId && canOpenDetails ? (
                <Button asChild variant='secondary' className='lg:self-stretch'>
                  <Link to={detailPath}>
                    Open schema details
                    <ExternalLink className='ml-2 h-4 w-4' />
                  </Link>
                </Button>
              ) : (
                <Button variant='secondary' disabled className='lg:self-stretch'>
                  {detailButtonLabel}
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
            <DatePickerWithRange
              aria-disabled={timeframeControlsDisabled}
              inert={timeframeControlsDisabled}
              className={timeframeControlsDisabled ? 'pointer-events-none opacity-60' : undefined}
              date={{
                from: startDate,
                to: deadline,
              }}
              setDate={(dateRange) => {
                if (timeframeControlsDisabled) {
                  return
                }

                onStartDateChange(dateRange?.from ? startOfDay(dateRange.from) : undefined)
                onDeadlineChange(dateRange?.to ? endOfDay(dateRange.to) : undefined)
              }}
            />
          </div>
        </div>
      )}

      {isActive && children}

      {isActive && (
        <div className='flex flex-wrap items-center gap-3 border-t border-border pt-4'>
          <p className='flex-1 text-xs leading-5 text-muted-foreground'>
            Save this card to apply its configuration changes for the current course phase.
          </p>
          <Button onClick={onSave} disabled={saveDisabled} className='ml-auto min-w-[160px]'>
            {isSaving ? 'Saving...' : `Save ${content.title}`}
          </Button>
        </div>
      )}
    </div>
  )

  if (!renderAsCard) {
    return contentBlock
  }

  return <Card className='border-border shadow-sm'>{contentBlock}</Card>
}
