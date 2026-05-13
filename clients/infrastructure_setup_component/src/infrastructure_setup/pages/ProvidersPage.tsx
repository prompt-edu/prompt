import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Server } from 'lucide-react'
import { infrastructureSetupAxiosInstance } from '../network/infrastructureSetupServerConfig'

// TODO: Define ProviderConfig interface once the API shape is known
interface ProviderConfig {
  id: string
  name: string
  type: string
}

const fetchProviderConfigs = async (coursePhaseID: string): Promise<ProviderConfig[]> => {
  const response = await infrastructureSetupAxiosInstance.get(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/provider-configs`,
  )
  return response.data
}

export const ProvidersPage = () => {
  const { coursePhaseID } = useParams<{ coursePhaseID: string }>()

  const {
    data: providers,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => fetchProviderConfigs(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  if (isLoading) {
    return <div className='p-4 text-gray-500'>Loading providers...</div>
  }

  if (isError) {
    return <div className='p-4 text-red-500'>Failed to load provider configurations.</div>
  }

  return (
    <div className='p-6 space-y-4'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center space-x-2'>
          <Server className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Providers</h1>
        </div>
        {/* TODO: Add button to open add/edit provider dialog */}
        <button className='px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm'>
          Add Provider
        </button>
      </div>

      {providers && providers.length === 0 ? (
        <div className='p-4 border-2 border-dashed border-gray-300 rounded-lg text-gray-500'>
          No provider configurations found. Add a provider to get started.
        </div>
      ) : (
        <div className='space-y-2'>
          {/* TODO: Render provider cards/rows */}
          {providers?.map((provider) => (
            <div
              key={provider.id}
              className='p-4 border border-gray-200 rounded-lg flex items-center justify-between'
            >
              <div>
                <p className='font-medium'>{provider.name}</p>
                <p className='text-sm text-gray-500'>{provider.type}</p>
              </div>
              {/* TODO: Add edit/delete actions */}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default ProvidersPage
