import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Button, cn, ErrorPage, Input, LoadingPage, useToast } from '@tumaet/prompt-ui-components'
import { Play, RefreshCw } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { InstanceRow } from '../components/InstanceRow'
import type { ResourceStatus } from '../interfaces/resourceInstance'
import { describeTriggerSummary } from '../interfaces/triggerSummary'
import { triggerExecution } from '../network/mutations/triggerExecution'
import { getInstances } from '../network/queries/getInstances'
import { describeError, hasStatus } from '../utils/describeError'

const isPollingStatus = (status: string) => status === 'pending' || status === 'in_progress'

const isConflict = (err: unknown) => hasStatus(err, 409)

const STATUS_ORDER: ResourceStatus[] = ['failed', 'partial', 'in_progress', 'pending', 'created']

export const ExecutionPage = () => {
  const { phaseId: coursePhaseID } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()
  const [statusFilter, setStatusFilter] = useState<ResourceStatus | 'all'>('all')
  const [search, setSearch] = useState('')

  const {
    data: instances,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['instances', coursePhaseID],
    queryFn: () => getInstances(coursePhaseID!),
    enabled: !!coursePhaseID,
    refetchInterval: (query) =>
      (query.state.data ?? []).some((i) => isPollingStatus(i.status)) ? 3000 : false,
  })

  const { mutate: execute, isPending: isExecuting } = useMutation({
    mutationFn: () => triggerExecution(coursePhaseID!),
    onSuccess: (summary) => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      const startedWork = summary.queued + summary.requeued > 0
      toast({
        title: startedWork ? 'Execution started' : 'Nothing left to provision',
        description: describeTriggerSummary(summary),
      })
    },
    onError: (err: unknown) => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      toast({
        title: isConflict(err) ? 'An execution is already running' : 'Failed to trigger execution',
        description: describeError(err),
        variant: 'destructive',
      })
    },
  })

  const hasRunningWork = (instances ?? []).some((i) => isPollingStatus(i.status))

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

  if (isLoading) {
    return <LoadingPage />
  }
  if (isError) {
    return <ErrorPage description='Failed to load execution instances.' onRetry={() => refetch()} />
  }

  return (
    <div className='space-y-4 p-6'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Play className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Execution</h1>
        </div>
        <div className='flex items-center gap-2'>
          <Button variant='outline' size='icon' onClick={() => refetch()} title='Refresh'>
            <RefreshCw className='h-4 w-4' />
          </Button>
          <Button
            onClick={() => execute()}
            disabled={isExecuting || hasRunningWork}
            title={hasRunningWork ? 'An execution is already running' : undefined}
          >
            <Play className='mr-2 h-4 w-4' />
            {isExecuting ? 'Triggering…' : 'Trigger execution'}
          </Button>
        </div>
      </div>

      {!instances || instances.length === 0 ? (
        <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
          No execution instances found. Trigger an execution to get started.
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
              No instance matches the current filter.
            </div>
          ) : (
            <div className='space-y-2'>
              {visible.map((instance) => (
                <InstanceRow key={instance.id} coursePhaseID={coursePhaseID!} instance={instance} />
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}

export default ExecutionPage
