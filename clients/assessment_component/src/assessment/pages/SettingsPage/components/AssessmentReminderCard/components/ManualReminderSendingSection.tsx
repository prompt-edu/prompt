import { Badge, Button } from '@tumaet/prompt-ui-components'

import type { EvaluationReminderType } from '../../../../../interfaces/evaluationReminder'
import { DeadlineBadge } from '../../../../components/badges'
import type { ReminderTypeConfig } from '../interfaces/ReminderTypeConfig'
import { formatDeadline, formatSentAt, mapReminderTypeToAssessmentType } from '../utils'

interface ManualReminderSendingSectionProps {
  reminderTypes: ReminderTypeConfig[]
  lastSentAtByType: Partial<Record<EvaluationReminderType, string>>
  getDisableReason: (config: ReminderTypeConfig) => string | undefined
  isSending: boolean
  onSend: (type: EvaluationReminderType) => void
}

export function ManualReminderSendingSection({
  reminderTypes,
  lastSentAtByType,
  getDisableReason,
  isSending,
  onSend,
}: ManualReminderSendingSectionProps) {
  return (
    <div className='space-y-4'>
      <div className='space-y-1'>
        <h3 className='text-base font-semibold text-foreground'>Send reminders</h3>
        <p className='text-sm leading-6 text-muted-foreground'>
          Choose the evaluation type. Recipients are selected automatically when the mail is sent.
        </p>
      </div>

      <div className='overflow-hidden rounded-lg border border-border'>
        {reminderTypes.map((reminderType) => {
          const lastSent = lastSentAtByType[reminderType.type]
          const disableReason = getDisableReason(reminderType)
          const disabled = !!disableReason || isSending

          return (
            <div
              key={reminderType.type}
              className='flex flex-col gap-4 border-b border-border bg-background p-4 last:border-b-0'
            >
              <div className='flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between'>
                <div className='min-w-0 space-y-3'>
                  <div className='flex flex-wrap items-center gap-2'>
                    <p className='font-medium leading-none'>{reminderType.label}</p>
                    {reminderType.deadline && (
                      <DeadlineBadge
                        deadline={reminderType.deadline}
                        type={mapReminderTypeToAssessmentType(reminderType.type)}
                      />
                    )}
                    <Badge variant={disabled ? 'outline' : 'secondary'}>
                      {disabled ? 'Not ready' : 'Ready to send'}
                    </Badge>
                  </div>

                  <div className='grid gap-1 text-sm text-muted-foreground sm:grid-cols-2'>
                    <p>Deadline: {formatDeadline(reminderType.deadline)}</p>
                    <p>Last sent: {formatSentAt(lastSent)}</p>
                  </div>

                  {disableReason && <p className='text-sm text-destructive'>{disableReason}</p>}
                </div>

                <Button
                  onClick={() => onSend(reminderType.type)}
                  disabled={disabled}
                  className='w-full lg:w-auto lg:min-w-[120px]'
                  size='sm'
                >
                  Send
                </Button>
              </div>
            </div>
          )
        })}
      </div>
    </div>
  )
}
