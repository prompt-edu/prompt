import { Badge, Button, Card, CardContent } from '@tumaet/prompt-ui-components'

import type { EvaluationReminderType } from '../../../../../interfaces/evaluationReminder'
import { DeadlineBadge } from '../../../../components/badges'
import type { ReminderTypeConfig } from '../interfaces/ReminderTypeConfig'
import { formatSentAt, mapReminderTypeToAssessmentType } from '../utils'

interface ManualReminderSendingSectionProps {
  reminderTypes: ReminderTypeConfig[]
  lastSentAtByType: Partial<Record<EvaluationReminderType, string>>
  getDisableReason: (config: ReminderTypeConfig) => string | undefined
  isEvaluationCompletionsPending: boolean
  isSending: boolean
  onSend: (type: EvaluationReminderType) => void
}

export function ManualReminderSendingSection({
  reminderTypes,
  lastSentAtByType,
  getDisableReason,
  isEvaluationCompletionsPending,
  isSending,
  onSend,
}: ManualReminderSendingSectionProps) {
  return (
    <div className='space-y-4'>
      <div className='space-y-1'>
        <h3 className='text-sm font-semibold text-foreground'>Manual Reminder Sending</h3>
      </div>
      <p className='text-sm leading-6 text-muted-foreground'>
        Only active evaluation types are shown. Reminders are sent to currently incomplete students
        only.
      </p>

      <div className='grid gap-3 lg:grid-cols-2 xl:grid-cols-3'>
        {reminderTypes.map((reminderType) => {
          const lastSent = lastSentAtByType[reminderType.type]
          const disableReason = getDisableReason(reminderType)
          const disabled = !!disableReason || isSending

          return (
            <Card key={reminderType.type} className='border-border shadow-sm'>
              <CardContent className='flex h-full flex-col gap-4 p-4'>
                <div className='flex items-start justify-between gap-3'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <p className='font-medium leading-none'>{reminderType.label}</p>
                    {reminderType.deadline && (
                      <DeadlineBadge
                        deadline={reminderType.deadline}
                        type={mapReminderTypeToAssessmentType(reminderType.type)}
                      />
                    )}
                  </div>
                  <Badge variant='secondary'>
                    {isEvaluationCompletionsPending
                      ? 'Calculating...'
                      : `${reminderType.recipientCount} ${
                          reminderType.recipientCount === 1 ? 'mail' : 'mails'
                        }`}
                  </Badge>
                </div>

                <div className='space-y-1 text-sm text-muted-foreground'>
                  <p>Last sent: {formatSentAt(lastSent)}</p>
                  {!isEvaluationCompletionsPending && (
                    <p>
                      Would send: {reminderType.recipientCount}{' '}
                      {reminderType.recipientCount === 1 ? 'mail' : 'mails'}
                    </p>
                  )}
                  {disableReason && <p className='text-destructive'>{disableReason}</p>}
                </div>

                <Button
                  onClick={() => onSend(reminderType.type)}
                  disabled={disabled}
                  className='mt-auto w-full'
                  size='sm'
                >
                  Send Reminder
                </Button>
              </CardContent>
            </Card>
          )
        })}
      </div>
    </div>
  )
}
