import { Clock } from 'lucide-react'

interface PrivacyExportRateLimitNoticeProps {
  until: string
  className?: string
}

export function PrivacyExportRateLimitNotice({
  until,
  className = '',
}: PrivacyExportRateLimitNoticeProps) {
  const displayDate = new Date(until).toLocaleDateString()
  return (
    <div
      className={
        'rounded-lg border border-amber-200 bg-amber-50 p-4 flex items-center gap-3 ' + className
      }
    >
      <Clock className='h-5 w-5 text-amber-600' />
      <div>
        <p className='font-semibold text-amber-900'>Export not available</p>
        <p className='text-sm text-amber-700'>You requested an export recently.</p>
        <p className='text-sm text-amber-700'>You can request a new one after {displayDate}</p>
      </div>
    </div>
  )
}
