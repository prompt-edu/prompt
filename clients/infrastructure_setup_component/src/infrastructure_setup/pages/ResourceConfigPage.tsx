import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Settings } from 'lucide-react'
import { infrastructureSetupAxiosInstance } from '../network/infrastructureSetupServerConfig'

// TODO: Define ResourceConfig interface once the API shape is known
interface ResourceConfig {
  id: string
  name: string
  providerConfigId: string
}

const fetchResourceConfigs = async (coursePhaseID: string): Promise<ResourceConfig[]> => {
  const response = await infrastructureSetupAxiosInstance.get(
    `infrastructure-setup/api/course_phase/${coursePhaseID}/resource-configs`,
  )
  return response.data
}

export const ResourceConfigPage = () => {
  const { coursePhaseID } = useParams<{ coursePhaseID: string }>()

  const {
    data: resourceConfigs,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['resource-configs', coursePhaseID],
    queryFn: () => fetchResourceConfigs(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  if (isLoading) {
    return <div className='p-4 text-gray-500'>Loading resource configurations...</div>
  }

  if (isError) {
    return <div className='p-4 text-red-500'>Failed to load resource configurations.</div>
  }

  return (
    <div className='p-6 space-y-4'>
      <div className='flex items-center space-x-2'>
        <Settings className='h-5 w-5 text-blue-500' />
        <h1 className='text-xl font-semibold'>Resource Configurations</h1>
      </div>

      {resourceConfigs && resourceConfigs.length === 0 ? (
        <div className='p-4 border-2 border-dashed border-gray-300 rounded-lg text-gray-500'>
          No resource configurations found.
        </div>
      ) : (
        <div className='space-y-2'>
          {/* TODO: Render resource config cards/rows */}
          {resourceConfigs?.map((config) => (
            <div key={config.id} className='p-4 border border-gray-200 rounded-lg'>
              <p className='font-medium'>{config.name}</p>
              <p className='text-sm text-gray-500'>Provider: {config.providerConfigId}</p>
              {/* TODO: Add edit/delete actions */}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default ResourceConfigPage
