import { Card, CardContent, ManagementPageHeader } from '@tumaet/prompt-ui-components'
import { ClipboardX } from 'lucide-react'

interface AssessmentDisabledNoticeProps {
  title: string
}

export const AssessmentDisabledNotice = ({ title }: AssessmentDisabledNoticeProps) => (
  <div className='space-y-4'>
    <ManagementPageHeader>{title}</ManagementPageHeader>
    <Card>
      <CardContent className='flex items-center gap-3 p-6'>
        <ClipboardX className='h-5 w-5 shrink-0 text-muted-foreground' />
        <p className='text-sm text-muted-foreground'>
          Assessment is disabled for this phase. It collects evaluations only. Enable the assessment
          in the phase settings to use this page.
        </p>
      </CardContent>
    </Card>
  </div>
)
