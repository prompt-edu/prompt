import { useQuery } from '@tanstack/react-query'
import { Button } from '@tumaet/prompt-ui-components'
import { PlusCircle, Settings } from 'lucide-react'
import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { ResourceConfigCard } from '../components/ResourceConfigCard'
import { ResourceConfigUpsertDialog } from '../dialogs/ResourceConfigUpsertDialog'
import type { ResourceConfig } from '../interfaces/resourceConfig'
import { getProviderConfigs } from '../network/queries/getProviderConfigs'
import { getResourceConfigs } from '../network/queries/getResourceConfigs'

export const ResourceConfigPage = () => {
  const { phaseId: coursePhaseID } = useParams<{ phaseId: string }>()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<ResourceConfig | undefined>(undefined)

  const {
    data: resourceConfigs,
    isLoading,
    isError,
  } = useQuery({
    queryKey: ['resource-configs', coursePhaseID],
    queryFn: () => getResourceConfigs(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  const { data: providers } = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => getProviderConfigs(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  if (isLoading) {
    return <div className='p-4 text-muted-foreground'>Loading resource configurations…</div>
  }
  if (isError) {
    return <div className='p-4 text-red-600'>Failed to load resource configurations.</div>
  }

  const openCreate = () => {
    setEditing(undefined)
    setDialogOpen(true)
  }

  const openEdit = (config: ResourceConfig) => {
    setEditing(config)
    setDialogOpen(true)
  }

  const availableProviderTypes = (providers ?? []).map((p) => p.providerType)

  return (
    <div className='space-y-4 p-6'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Settings className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Resource configurations</h1>
        </div>
        <Button onClick={openCreate} disabled={availableProviderTypes.length === 0}>
          <PlusCircle className='mr-2 h-4 w-4' />
          New resource config
        </Button>
      </div>

      {availableProviderTypes.length === 0 && (
        <div className='rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900'>
          Add at least one provider before creating resource configurations.
        </div>
      )}

      {!resourceConfigs || resourceConfigs.length === 0 ? (
        <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
          No resource configurations found.
        </div>
      ) : (
        <div className='space-y-2'>
          {resourceConfigs.map((config) => (
            <ResourceConfigCard
              key={config.id}
              coursePhaseID={coursePhaseID!}
              config={config}
              onEdit={openEdit}
            />
          ))}
        </div>
      )}

      {coursePhaseID && (
        <ResourceConfigUpsertDialog
          coursePhaseID={coursePhaseID}
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          existing={editing}
          availableProviderTypes={availableProviderTypes}
        />
      )}
    </div>
  )
}

export default ResourceConfigPage
