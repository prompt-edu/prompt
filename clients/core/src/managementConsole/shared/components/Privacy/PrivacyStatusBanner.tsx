import type { ReactNode } from 'react'

interface PrivacyStatusBannerProps {
  icon: ReactNode
  title: string
  meta?: ReactNode[]
  action?: ReactNode
  footer?: ReactNode
}

export function PrivacyStatusBanner({
  icon,
  title,
  meta,
  action,
  footer,
}: PrivacyStatusBannerProps) {
  return (
    <div className='rounded-lg border border-border bg-muted p-4 flex flex-col gap-3'>
      <div className='flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4'>
        <div className='flex items-center gap-3'>
          <div className='lg:mx-2'>{icon}</div>
          <div className='flex-1'>
            <p className='font-semibold text-foreground'>{title}</p>
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
