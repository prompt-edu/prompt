import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Play, RefreshCw } from 'lucide-react'
import { infrastructureSetupAxiosInstance } from '../network/infrastructureSetupServerConfig'

// TODO: Define Instance interface once the API shape is known
interface Instance {
  id: string
  status: string
  createdAt: string
}

const fetchInstances = async (coursePhaseID: string): Promise<Instance[]> => {
  const response = await infrastructureSetupAxiosInstance.get(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/instances`,
  )
  return response.data
}

const triggerExecution = async (coursePhaseID: string): Promise<void> => {
  await infrastructureSetupAxiosInstance.post(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/execute`,
  )
}

export const ExecutionPage = () => {
  const { coursePhaseID } = useParams<{ coursePhaseID: string }>()
  const queryClient = useQueryClient()

  const {
    data: instances,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['instances', coursePhaseID],
    queryFn: () => fetchInstances(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  const { mutate: execute, isPending: isExecuting } = useMutation({
    mutationFn: () => triggerExecution(coursePhaseID!),
    onSuccess: () => {
      // Refresh instances after triggering execution
      queryClient.invalidateQueries({ queryKey: ['instances', coursePhaseID] })
    },
    // TODO: Add error handling / toast notification
  })

  if (isLoading) {
    return <div className='p-4 text-gray-500'>Loading execution instances...</div>
  }

  if (isError) {
    return <div className='p-4 text-red-500'>Failed to load execution instances.</div>
  }

  return (
    <div className='p-6 space-y-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <Play className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Execution</h1>
        </div>
        <div className='flex items-center space-x-2'>
          <button
            onClick={() => refetch()}
            className='p-2 border border-gray-300 rounded hover:bg-gray-100'
            title='Refresh'
          >
            <RefreshCw className='h-4 w-4' />
          </button>
          <button
            onClick={() => execute()}
            disabled={isExecuting}
            className='px-4 py-2 bg-green-600 text-white rounded hover:bg-green-700 text-sm disabled:opacity-50 flex items-center space-x-1'
          >
            <Play className='h-4 w-4' />
            <span>{isExecuting ? 'Triggering...' : 'Trigger Execution'}</span>
          </button>
        </div>
      </div>

      {instances && instances.length === 0 ? (
        <div className='p-4 border-2 border-dashed border-gray-300 rounded-lg text-gray-500'>
          No execution instances found. Trigger an execution to get started.
        </div>
      ) : (
        <div className='space-y-2'>
          {/* TODO: Render instance cards/rows with status indicators */}
          {instances?.map((instance) => (
            <div
              key={instance.id}
              className='p-4 border border-gray-200 rounded-lg flex items-center justify-between'
            >
              <div>
                <p className='font-medium text-sm font-mono'>{instance.id}</p>
                <p className='text-sm text-gray-500'>{instance.createdAt}</p>
              </div>
              <span className='px-2 py-1 bg-gray-100 text-gray-700 rounded text-xs'>
                {instance.status}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default ExecutionPage
