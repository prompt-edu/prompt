import { Badge } from '@tumaet/prompt-ui-components'

import type { ResourceStatus } from '../interfaces/resourceInstance'

const statusToVariant: Record<ResourceStatus, 'default' | 'secondary' | 'outline' | 'destructive'> =
  {
    pending: 'outline',
    in_progress: 'secondary',
    created: 'default',
    failed: 'destructive',
  }

const statusToLabel: Record<ResourceStatus, string> = {
  pending: 'pending',
  in_progress: 'in progress',
  created: 'created',
  failed: 'failed',
}

export const StatusBadge = ({ status }: { status: ResourceStatus }) => (
  <Badge variant={statusToVariant[status]} className='capitalize'>
    {statusToLabel[status]}
  </Badge>
)

export default StatusBadge
