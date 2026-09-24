import { useQuery } from '@tanstack/react-query'
import { Button, Skeleton } from '@tumaet/prompt-ui-components'
import { PlusCircle } from 'lucide-react'
import { useState } from 'react'
import { ProviderUpsertDialog } from '../../dialogs/ProviderUpsertDialog'
import type { ProviderConfig } from '../../interfaces/providerConfig'
import { getProviderConfigs } from '../../network/queries/getProviderConfigs'
import { ProviderCard } from '../ProviderCard'
import { SectionError } from '../SectionError'
import { ConfigurationSection } from './ConfigurationSection'

interface Props {
  coursePhaseID: string
}

export const ProvidersSection = ({ coursePhaseID }: Props) => {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editingProvider, setEditingProvider] = useState<ProviderConfig | undefined>(undefined)

  const {
    data: providers,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => getProviderConfigs(coursePhaseID),
  })

  const openCreate = () => {
    setEditingProvider(undefined)
    setDialogOpen(true)
  }

  const openEdit = (provider: ProviderConfig) => {
    setEditingProvider(provider)
    setDialogOpen(true)
  }

  const configuredTypes = (providers ?? []).map((p) => p.providerType)

  return (
    <ConfigurationSection
      id='providers'
      step={2}
      title='Providers'
      description='The external systems resources are created in, with the credentials PROMPT uses for them. Credentials are stored encrypted and never shown again.'
      action={
        <Button onClick={openCreate} disabled={isLoading || isError}>
          <PlusCircle className='mr-2 h-4 w-4' />
          Add provider
        </Button>
      }
    >
      {isLoading ? (
        <Skeleton className='h-16 w-full' />
      ) : isError ? (
        <SectionError message='Failed to load the providers.' onRetry={() => refetch()} />
      ) : !providers || providers.length === 0 ? (
        <div className='rounded-lg border-2 border-dashed border-gray-300 p-4 text-muted-foreground'>
          No provider configured yet. Add one to start.
        </div>
      ) : (
        <div className='space-y-2'>
          {providers.map((provider) => (
            <ProviderCard
              key={provider.id}
              coursePhaseID={coursePhaseID}
              provider={provider}
              onEdit={openEdit}
            />
          ))}
        </div>
      )}

      <ProviderUpsertDialog
        coursePhaseID={coursePhaseID}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        existingProvider={editingProvider}
        configuredTypes={configuredTypes}
      />
    </ConfigurationSection>
  )
}
