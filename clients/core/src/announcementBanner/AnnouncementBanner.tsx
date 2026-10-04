import type { AnnouncementSeverity } from '@core/interfaces/announcement'
import { Button, cn } from '@tumaet/prompt-ui-components'
import { ChevronLeft, ChevronRight, Info, OctagonAlert, TriangleAlert, X } from 'lucide-react'

interface AnnouncementPager {
  position: number
  total: number
  onPrevious: () => void
  onNext: () => void
}

interface AnnouncementBannerProps {
  severity: AnnouncementSeverity
  title: string
  message: string
  linkUrl: string
  linkLabel: string
  pager?: AnnouncementPager
  onDismiss?: () => void
  className?: string
}

const SEVERITY_STYLES: Record<AnnouncementSeverity, string> = {
  info: 'bg-blue-600 text-white dark:bg-blue-700',
  warning: 'bg-amber-300 text-amber-950 dark:bg-amber-400',
  critical: 'bg-red-600 text-white dark:bg-red-700',
}

const SEVERITY_ICONS: Record<AnnouncementSeverity, typeof Info> = {
  info: Info,
  warning: TriangleAlert,
  critical: OctagonAlert,
}

const SEVERITY_LABELS: Record<AnnouncementSeverity, string> = {
  info: 'Information',
  warning: 'Warning',
  critical: 'Critical',
}

const ICON_BUTTON_CLASSES =
  'h-7 w-7 shrink-0 text-current hover:bg-black/15 hover:text-current disabled:opacity-100'

export const AnnouncementBanner = ({
  severity,
  title,
  message,
  linkUrl,
  linkLabel,
  pager,
  onDismiss,
  className,
}: AnnouncementBannerProps) => {
  const Icon = SEVERITY_ICONS[severity]

  return (
    <div
      role={severity === 'critical' ? 'alert' : 'status'}
      className={cn('w-full text-sm', SEVERITY_STYLES[severity], className)}
    >
      <div className='grid min-h-11 grid-cols-[1fr_minmax(0,auto)_1fr] items-center gap-3 px-4 py-1.5'>
        <span aria-hidden='true' />
        <div className='flex min-w-0 items-center justify-center gap-3 text-center'>
          <Icon className='h-4 w-4 shrink-0' aria-hidden='true' />
          <p className='min-w-0 wrap-break-word'>
            <span className='sr-only'>{SEVERITY_LABELS[severity]}: </span>
            {title && <span className='mr-2 font-semibold'>{title}</span>}
            <span>{message}</span>
            {linkUrl && (
              <a
                href={linkUrl}
                target='_blank'
                rel='noopener noreferrer'
                className='ml-2 whitespace-nowrap font-semibold underline underline-offset-2 hover:no-underline'
              >
                {linkLabel || 'Learn more'}
              </a>
            )}
          </p>
        </div>
        <div className='flex shrink-0 items-center justify-self-end'>
          {pager && (
            <div className='flex items-center'>
              <Button
                variant='ghost'
                size='icon'
                className={ICON_BUTTON_CLASSES}
                onClick={pager.onPrevious}
                aria-label='Previous announcement'
              >
                <ChevronLeft className='h-4 w-4' />
              </Button>
              <span className='min-w-8 text-center text-xs font-medium tabular-nums'>
                {pager.position}/{pager.total}
              </span>
              <Button
                variant='ghost'
                size='icon'
                className={ICON_BUTTON_CLASSES}
                onClick={pager.onNext}
                aria-label='Next announcement'
              >
                <ChevronRight className='h-4 w-4' />
              </Button>
            </div>
          )}
          <Button
            variant='ghost'
            size='icon'
            className={ICON_BUTTON_CLASSES}
            onClick={onDismiss}
            disabled={!onDismiss}
            aria-label='Dismiss announcement'
          >
            <X className='h-4 w-4' />
          </Button>
        </div>
      </div>
    </div>
  )
}
