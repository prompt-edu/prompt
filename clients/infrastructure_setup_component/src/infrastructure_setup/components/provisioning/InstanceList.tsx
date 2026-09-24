import { Button, cn, Input, Skeleton } from '@tumaet/prompt-ui-components'
import { RefreshCw } from 'lucide-react'
import { useState } from 'react'
import type { ResourceInstance, ResourceStatus } from '../../interfaces/resourceInstance'
import { InstanceRow } from '../InstanceRow'
import { SectionError } from '../SectionError'

interface Props {
  coursePhaseID: string
  instances: ResourceInstance[] | undefined
  isLoading: boolean
  isError: boolean
  onRefresh: () => void
}

const STATUS_ORDER: ResourceStatus[] = ['failed', 'partial', 'in_progress', 'pending', 'created']

export const InstanceList = ({
  coursePhaseID,
  instances,
  isLoading,
  isError,
  onRefresh,
}: Props) => {
  const [statusFilter, setStatusFilter] = useState<ResourceStatus | 'all'>('all')
  const [search, setSearch] = useState('')

  const counts = STATUS_ORDER.reduce<Record<ResourceStatus, number>>(
    (acc, status) => {
      acc[status] = (instances ?? []).filter((i) => i.status === status).length
      return acc
    },
    { pending: 0, in_progress: 0, created: 0, partial: 0, failed: 0 },
  )

  const needle = search.trim().toLowerCase()
  const visible = (instances ?? []).filter(
    (instance) =>
      (statusFilter === 'all' || instance.status === statusFilter) &&
      (needle === '' ||
        instance.targetName.toLowerCase().includes(needle) ||
        instance.resolvedName.toLowerCase().includes(needle) ||
        instance.nameTemplate.toLowerCase().includes(needle)),
  )

  return (
    <section className='space-y-3'>
      <div className='flex items-center justify-between'>
        <div className='space-y-1'>
          <h2 className='text-xl font-semibold'>Status</h2>
          <p className='text-sm text-muted-foreground'>
            One row per resource and team or student. Failed and partial rows can be retried one by
            one, or all at once with the next run.
          </p>
        </div>
        <Button variant='outline' size='icon' onClick={onRefresh} title='Refresh'>
          <RefreshCw className='h-4 w-4' />
        </Button>
      </div>

      {isLoading ? (
        <Skeleton className='h-24 w-full' />
      ) : isError ? (
        <SectionError message='Failed to load the provisioned resources.' onRetry={onRefresh} />
      ) : !instances || instances.length === 0 ? (
        <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
          Nothing has been provisioned yet.
        </div>
      ) : (
        <>
          <div className='flex flex-wrap items-center gap-2'>
            {STATUS_ORDER.filter((status) => counts[status] > 0).map((status) => (
              <button
                key={status}
                type='button'
                onClick={() => setStatusFilter(statusFilter === status ? 'all' : status)}
                className={cn(
                  'rounded-full border px-3 py-1 text-sm',
                  statusFilter === status
                    ? 'border-blue-500 bg-blue-50 text-blue-900'
                    : 'border-transparent bg-muted',
                )}
              >
                {counts[status]} {status.replace('_', ' ')}
              </button>
            ))}
            <Input
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder='Filter by team, student or resource name'
              className='ml-auto max-w-xs'
            />
          </div>

          {visible.length === 0 ? (
            <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
              No resource matches the current filter.
            </div>
          ) : (
            <div className='space-y-2'>
              {visible.map((instance) => (
                <InstanceRow key={instance.id} coursePhaseID={coursePhaseID} instance={instance} />
              ))}
            </div>
          )}
        </>
      )}
    </section>
  )
}
