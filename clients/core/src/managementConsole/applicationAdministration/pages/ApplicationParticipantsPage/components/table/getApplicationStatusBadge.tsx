import { PassStatus } from '@tumaet/prompt-shared-state'
import { Badge } from '@tumaet/prompt-ui-components'

interface ApplicationStatusDisplayConfig {
  label: string
  badgeClassName: string
}

const defaultStatusDisplayConfig: ApplicationStatusDisplayConfig = {
  label: 'Unknown',
  badgeClassName: 'bg-gray-500 hover:bg-gray-500',
}

export const APPLICATION_PASS_STATUS_LABELS = {
  [PassStatus.PASSED]: 'Accepted',
  [PassStatus.FAILED]: 'Rejected',
}

const applicationStatusDisplayConfig: Record<PassStatus, ApplicationStatusDisplayConfig> = {
  [PassStatus.PASSED]: {
    label: APPLICATION_PASS_STATUS_LABELS[PassStatus.PASSED],
    badgeClassName: 'bg-green-500 hover:bg-green-500',
  },
  [PassStatus.FAILED]: {
    label: APPLICATION_PASS_STATUS_LABELS[PassStatus.FAILED],
    badgeClassName: 'bg-red-500 hover:bg-red-500',
  },
  [PassStatus.NOT_ASSESSED]: {
    label: 'Not Assessed',
    badgeClassName: 'bg-gray-500 hover:bg-gray-500',
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
