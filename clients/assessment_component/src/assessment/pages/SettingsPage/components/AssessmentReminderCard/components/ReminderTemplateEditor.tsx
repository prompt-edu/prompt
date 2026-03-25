import type { RefObject } from 'react'

import { Button, Input, Label, Textarea } from '@tumaet/prompt-ui-components'

interface ReminderTemplateEditorProps {
  subject: string
  content: string
  onSubjectChange: (value: string) => void
  onContentChange: (value: string) => void
  onSave: () => void
  isPending: boolean
  isSaving: boolean
  isModified: boolean
  contentTextareaRef: RefObject<HTMLTextAreaElement | null>
}

export function ReminderTemplateEditor({
  subject,
  content,
  onSubjectChange,
  onContentChange,
  onSave,
  isPending,
  isSaving,
  isModified,
  contentTextareaRef,
}: ReminderTemplateEditorProps) {
  return (
    <div className='space-y-6'>
      <div className='space-y-3'>
        <div className='space-y-1'>
          <h3 className='text-sm font-semibold text-foreground'>Reminder Template</h3>
          <p className='text-sm leading-6 text-muted-foreground'>
            This shared template is used for all manual reminder sends.
          </p>
        </div>

        <div className='space-y-3'>
          <Label htmlFor='assessment-reminder-subject'>Reminder Subject</Label>
          <Input
            id='assessment-reminder-subject'
            placeholder='Assessment reminder for {{evaluationType}}'
            value={subject}
            onChange={(event) => onSubjectChange(event.target.value)}
            disabled={isPending}
          />
        </div>

        <div className='space-y-3'>
          <Label htmlFor='assessment-reminder-content'>Reminder Message</Label>
          <Textarea
            ref={contentTextareaRef}
            id='assessment-reminder-content'
            placeholder='Hi {{firstName}}, please complete your {{evaluationType}} before {{evaluationDeadline}}.'
            className='min-h-[130px] resize-none overflow-hidden'
            value={content}
            onChange={(event) => onContentChange(event.target.value)}
            disabled={isPending}
          />
        </div>
      </div>

      <div className='flex flex-wrap items-center gap-3 border-t border-border pt-4'>
        <p className='flex-1 text-xs leading-5 text-muted-foreground'>
          Save this section to apply reminder template changes for the current course phase.
        </p>
        <Button
          onClick={onSave}
          disabled={!isModified || isSaving || isPending}
          className='min-w-[160px]'
        >
          {isSaving ? 'Saving...' : 'Save Template'}
        </Button>
      </div>
    </div>
  )
}
