import { Badge } from '@tumaet/prompt-ui-components'
import { XCircle, AlertCircle, CheckCircle2 } from 'lucide-react'

interface ServiceStatusBadgeProps {
  status: 'Offline' | 'OnlineUnhealthy' | 'Online'
}

export function ServiceStatusBadge({ status }: ServiceStatusBadgeProps) {
  if (status == 'Offline') {
    return (
      <Badge variant='destructive' className='flex items-center gap-1'>
        <XCircle className='h-3 w-3' />
        Offline
      </Badge>
    )
  }
  if (status == 'OnlineUnhealthy') {
    return (
      <Badge
        variant='outline'
        className='flex items-center gap-1 border-yellow-500 text-yellow-600'
      >
        <AlertCircle className='h-3 w-3' />
        Degraded
      </Badge>
    )
  }
  return (
    <Badge variant='outline' className='flex items-center gap-1 border-green-500 text-green-600'>
      <CheckCircle2 className='h-3 w-3' />
      Online
    </Badge>
  )
}
