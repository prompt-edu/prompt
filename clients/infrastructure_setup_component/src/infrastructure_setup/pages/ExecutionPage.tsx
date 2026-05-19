import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Button, ErrorPage, LoadingPage, useToast } from '@tumaet/prompt-ui-components'
import { Play, RefreshCw } from 'lucide-react'

import { getInstances } from '../network/queries/getInstances'
import { triggerExecution } from '../network/mutations/triggerExecution'
import { InstanceRow } from '../components/InstanceRow'

const isPollingStatus = (status: string) => status === 'pending' || status === 'in_progress'

export const ExecutionPage = () => {
  const { phaseId: coursePhaseID } = useParams<{ phaseId: string }>()
  const queryClient = useQueryClient()
  const { toast } = useToast()

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
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
      toast({ title: 'Execution started' })
    },
    onError: (err: unknown) => {
      toast({
        title: 'Failed to trigger execution',
        description: err instanceof Error ? err.message : 'Unknown error',
        variant: 'destructive',
      })
    },
  })

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
          <Button onClick={() => execute()} disabled={isExecuting}>
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
        <div className='space-y-2'>
          {instances.map((instance) => (
            <InstanceRow key={instance.id} coursePhaseID={coursePhaseID!} instance={instance} />
          ))}
        </div>
      )}
    </div>
  )
}

export default ExecutionPage
