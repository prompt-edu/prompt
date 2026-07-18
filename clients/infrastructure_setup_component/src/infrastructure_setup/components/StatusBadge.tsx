import { Badge } from '@tumaet/prompt-ui-components'

import type { ResourceStatus } from '../interfaces/resourceInstance'

const statusToVariant: Record<ResourceStatus, 'default' | 'secondary' | 'outline' | 'destructive'> =
  {
    pending: 'outline',
    in_progress: 'secondary',
    created: 'default',
    failed: 'destructive',
  }

export const StatusBadge = ({ status }: { status: ResourceStatus }) => (
  <Badge variant={statusToVariant[status]} className='capitalize'>
    {status.replace('_', ' ')}
  </Badge>
)

export default StatusBadge
