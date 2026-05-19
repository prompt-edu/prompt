import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useParams } from 'react-router-dom'
import { Button, ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
import { PlusCircle, Server } from 'lucide-react'

import { ProviderConfig } from '../interfaces/providerConfig'
import { getProviderConfigs } from '../network/queries/getProviderConfigs'
import { ProviderCard } from '../components/ProviderCard'
import { ProviderUpsertDialog } from '../dialogs/ProviderUpsertDialog'

export const ProvidersPage = () => {
  const { phaseId: coursePhaseID } = useParams<{ phaseId: string }>()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<ProviderConfig | undefined>(undefined)

  const {
    data: providers,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => getProviderConfigs(coursePhaseID!),
    enabled: !!coursePhaseID,
  })

  const openCreate = () => {
    setEditingProvider(undefined)
    setDialogOpen(true)
  }

  const openEdit = (provider: ProviderConfig) => {
    setEditingProvider(provider)
    setDialogOpen(true)
  }

  if (isLoading) {
    return <LoadingPage />
  }
  if (isError) {
    return (
      <ErrorPage
        description='Failed to load provider configurations.'
        onRetry={() => refetch()}
      />
    )
  }

  const configuredTypes = (providers ?? []).map((p) => p.providerType)

  return (
    <div className='space-y-4 p-6'>
      <div className='flex items-center justify-between'>
        <div className='flex items-center gap-2'>
          <Server className='h-5 w-5 text-blue-500' />
          <h1 className='text-xl font-semibold'>Providers</h1>
        </div>
        <Button onClick={openCreate}>
          <PlusCircle className='mr-2 h-4 w-4' />
          Add provider
        </Button>
      </div>

      {!providers || providers.length === 0 ? (
        <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
          No provider configurations found. Add a provider to get started.
        </div>
      ) : (
        <div className='space-y-2'>
          {providers.map((provider) => (
            <ProviderCard
              key={provider.id}
              coursePhaseID={coursePhaseID!}
              provider={provider}
              onEdit={openEdit}
            />
          ))}
        </div>
      )}

      {coursePhaseID && (
        <ProviderUpsertDialog
          coursePhaseID={coursePhaseID}
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          existingProvider={editingProvider}
          configuredTypes={configuredTypes}
        />
      )}
    </div>
  )
}

export default ProvidersPage
