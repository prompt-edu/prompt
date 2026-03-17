import { Badge } from '@tumaet/prompt-ui-components'
import { PassStatus } from '@tumaet/prompt-shared-state'

interface ApplicationStatusDisplayConfig {
  label: string
  badgeClassName: string
  navigationButtonClassName: string
}

const defaultStatusDisplayConfig: ApplicationStatusDisplayConfig = {
  label: 'Unknown',
  badgeClassName: 'bg-gray-500 hover:bg-gray-500',
  navigationButtonClassName: 'border-gray-500 text-gray-700 hover:bg-gray-50',
}

const applicationStatusDisplayConfig: Record<PassStatus, ApplicationStatusDisplayConfig> = {
  [PassStatus.PASSED]: {
    label: 'Accepted',
    badgeClassName: 'bg-green-500 hover:bg-green-500',
    navigationButtonClassName: 'border-green-600 text-green-700 hover:bg-green-50',
  },
  [PassStatus.FAILED]: {
    label: 'Rejected',
    badgeClassName: 'bg-red-500 hover:bg-red-500',
    navigationButtonClassName: 'border-red-600 text-red-700 hover:bg-red-50',
  },
  [PassStatus.NOT_ASSESSED]: {
    label: 'Not Assessed',
    badgeClassName: 'bg-gray-500 hover:bg-gray-500',
    navigationButtonClassName: 'border-gray-500 text-gray-700 hover:bg-gray-50',
  },
}

function getApplicationStatusDisplayConfig(
  status: PassStatus | undefined,
): ApplicationStatusDisplayConfig {
  return (status && applicationStatusDisplayConfig[status]) || defaultStatusDisplayConfig
}

export function getApplicationStatusBadge(status: PassStatus) {
  const { label, badgeClassName } = getApplicationStatusDisplayConfig(status)
  return <Badge className={badgeClassName}>{label}</Badge>
}

export function getApplicationStatusString(status: PassStatus): string {
  return getApplicationStatusDisplayConfig(status).label
}

export function getApplicationNavigationButtonColorClass(status: PassStatus | undefined): string {
  return getApplicationStatusDisplayConfig(status).navigationButtonClassName
}
