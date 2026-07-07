import { Ban, CircleCheck, CircleX, Clock, Loader2 } from 'lucide-react'
import type { ReactNode } from 'react'

export type PrivacyStatusBannerState =
  | 'success'
  | 'partial'
  | 'failure'
  | 'in_progress'
  | 'pending_approval'
  | 'rejected'

interface PrivacyStatusBannerProps {
  subject: string
  state: PrivacyStatusBannerState
  meta?: ReactNode[]
  action?: ReactNode
  footer?: ReactNode
}

function StatusIcon({ state }: { state: PrivacyStatusBannerState }) {
  const className = 'h-5 w-5'
  switch (state) {
    case 'in_progress':
      return <Loader2 className={`${className} animate-spin text-muted-foreground`} />
    case 'pending_approval':
      return <Clock className={`${className} text-muted-foreground`} />
    case 'success':
      return <CircleCheck className={`${className} text-green-600 dark:text-green-400`} />
    case 'partial':
      return <CircleCheck className={`${className} text-amber-600 dark:text-amber-400`} />
    case 'failure':
      return <CircleX className={`${className} text-red-600 dark:text-red-400`} />
    case 'rejected':
      return <Ban className={`${className} text-red-600 dark:text-red-400`} />
  }
}

function getTitle(subject: string, state: PrivacyStatusBannerState): string {
  switch (state) {
    case 'in_progress':
      return `${subject} in progress`
    case 'pending_approval':
      return `${subject} pending approval`
    case 'success':
      return `${subject} completed`
    case 'partial':
      return `${subject} completed with issues`
    case 'failure':
      return `${subject} failed`
    case 'rejected':
      return `${subject} rejected`
  }
}

export function PrivacyStatusBanner({
  subject,
  state,
  meta,
  action,
  footer,
}: PrivacyStatusBannerProps) {
  return (
    <div className='rounded-lg border border-border bg-muted p-4 flex flex-col gap-3'>
      <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4'>
        <div className='flex items-center gap-3'>
          <div className='lg:mx-2'>
            <StatusIcon state={state} />
          </div>
          <div className='flex-1'>
            <p className='font-semibold text-foreground'>{getTitle(subject, state)}</p>
            {meta?.filter(Boolean).map((line, i) => (
              <p key={i} className='text-xs text-muted-foreground mt-0.5'>
                {line}
              </p>
            ))}
          </div>
        </div>
        {action}
      </div>
      {footer}
    </div>
  )
}
