import { Badge, cn } from '@tumaet/prompt-ui-components'

import type { ResourceState } from '../utils/resourceState'

const STATE_CLASS_NAMES: Record<ResourceState, string> = {
  ready: 'border-transparent bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300',
  partly_ready: 'border-amber-500 text-amber-700 dark:text-amber-400',
  not_added:
    'border-transparent bg-amber-100 text-amber-800 dark:bg-amber-900/40 dark:text-amber-300',
  setting_up: 'border-transparent bg-blue-100 text-blue-800 dark:bg-blue-900/40 dark:text-blue-300',
  failed: 'border-transparent bg-red-100 text-red-800 dark:bg-red-900/40 dark:text-red-300',
  not_provisioned: 'text-muted-foreground',
}

interface Props {
  state: ResourceState
  label: string
}

export const ResourceStateBadge = ({ state, label }: Props) => (
  <Badge variant='outline' className={cn('whitespace-nowrap', STATE_CLASS_NAMES[state])}>
    {label}
  </Badge>
)
