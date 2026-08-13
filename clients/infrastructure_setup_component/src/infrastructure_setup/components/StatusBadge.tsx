import { Badge, cn } from '@tumaet/prompt-ui-components'

import type { ResourceStatus } from '../interfaces/resourceInstance'

const statusToVariant: Record<ResourceStatus, 'default' | 'secondary' | 'outline' | 'destructive'> =
  {
    pending: 'outline',
    in_progress: 'secondary',
    created: 'default',
    partial: 'outline',
    failed: 'destructive',
  }

const statusToClassName: Partial<Record<ResourceStatus, string>> = {
  partial: 'border-amber-500 text-amber-600 dark:text-amber-400',
}

export const StatusBadge = ({ status }: { status: ResourceStatus }) => (
  <Badge variant={statusToVariant[status]} className={cn('capitalize', statusToClassName[status])}>
    {status.replace('_', ' ')}
  </Badge>
)

export default StatusBadge
