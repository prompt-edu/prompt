import { Badge } from '@tumaet/prompt-ui-components'
import { AlertCircle, CheckCircle2, Clock, XCircle } from 'lucide-react'
import { ApplicationStatus } from '../interfaces/applicationStatus'

interface ApplicationConfigurationHeaderProps {
  applicationPhaseIsConfigured: boolean
  applicationStatus: string
}

export const ApplicationStatusBadge = ({
  applicationPhaseIsConfigured,
  applicationStatus,
}: ApplicationConfigurationHeaderProps) => {
  const getStatusIcon = (status: string) => {
    switch (status) {
      case ApplicationStatus.NotConfigured:
        return <AlertCircle className='h-5 w-5 text-yellow-500' />
      case ApplicationStatus.NotYetLive:
        return <Clock className='h-5 w-5 text-blue-500' />
      case ApplicationStatus.Live:
        return <CheckCircle2 className='h-5 w-5 text-green-500' />
      case ApplicationStatus.Passed:
        return <XCircle className='h-5 w-5 text-gray-500' />
      default:
        return null
    }
  }

  return (
    <Badge
      variant={applicationPhaseIsConfigured ? 'default' : 'secondary'}
      className={`flex items-center space-x-1 ${
        applicationStatus === ApplicationStatus.Live
          ? 'bg-green-100 text-green-800 hover:bg-green-200 hover:text-green-800'
          : applicationStatus === ApplicationStatus.NotYetLive
            ? 'bg-blue-100 text-blue-800 hover:bg-blue-200 hover:text-blue-800'
            : applicationStatus === ApplicationStatus.Passed
              ? 'bg-gray-100 text-gray-800 hover:bg-gray-200 hover:text-gray-800'
              : 'bg-yellow-100 text-yellow-800 hover:bg-yellow-200 hover:text-yellow-800'
      }`}
    >
      {getStatusIcon(applicationStatus)}
      <span>{applicationStatus}</span>
    </Badge>
  )
}
