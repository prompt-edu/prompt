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
}: ReminderTemplateEditorProps) {
  return (
    <div className='space-y-5'>
      <div className='space-y-1'>
        <h3 className='text-base font-semibold text-foreground'>Reminder template</h3>
        <p className='text-sm leading-6 text-muted-foreground'>
          One shared template is used for self, peer, and tutor evaluation reminders.
        </p>
      </div>

      <div className='space-y-3'>
        <Label htmlFor='assessment-reminder-subject'>Subject</Label>
        <Input
          id='assessment-reminder-subject'
          placeholder='Assessment reminder for {{evaluationType}}'
          value={subject}
          onChange={(event) => onSubjectChange(event.target.value)}
          disabled={isPending}
        />
      </div>

      <div className='space-y-3'>
        <Label htmlFor='assessment-reminder-content'>Message</Label>
        <Textarea
          id='assessment-reminder-content'
          placeholder='Hi {{firstName}}, please complete your {{evaluationType}} before {{evaluationDeadline}}.'
          className='min-h-[180px] resize-y'
          value={content}
          onChange={(event) => onContentChange(event.target.value)}
          disabled={isPending}
        />
        <p className='text-xs leading-5 text-muted-foreground'>
          Keep the message general so it works for every reminder type.
        </p>
      </div>

      <div className='flex flex-wrap items-center gap-3 border-t border-border pt-4'>
        <p className='flex-1 text-xs leading-5 text-muted-foreground'>
          Save the template before sending reminders.
        </p>
        <Button
          onClick={onSave}
          disabled={!isModified || isSaving || isPending}
          className='min-w-[160px]'
        >
          {isSaving ? 'Saving...' : 'Save template'}
        </Button>
      </div>
    </div>
  )
}
