import { useQuery } from '@tanstack/react-query'
import { Button, ErrorPage, LoadingPage } from '@tumaet/prompt-ui-components'
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
    refetch,
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
    return <LoadingPage />
  }
  if (isError) {
    return (
      <ErrorPage description='Failed to load resource configurations.' onRetry={() => refetch()} />
    )
  }

  const openCreate = () => {
    setEditing(undefined)
    setDialogOpen(true)
  }

  const openEdit = (config: ResourceConfig) => {
    setEditing(config)
    setDialogOpen(true)
  }

  // Only providers holding credentials can back a resource config; the server rejects
  // the rest, so they are not offered here either.
  const availableProviderTypes = (providers ?? [])
    .filter((p) => p.configured)
    .map((p) => p.providerType)

  const outlineConfigs = (resourceConfigs ?? []).filter((c) => c.providerType === 'outline')
  const keycloakScopes = new Set(
    (resourceConfigs ?? []).filter((c) => c.providerType === 'keycloak').map((c) => c.scope),
  )
  const outlineWithoutKeycloak = outlineConfigs.filter((c) => !keycloakScopes.has(c.scope))

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
          {(providers ?? []).length === 0
            ? 'Add at least one provider before creating resource configurations.'
            : 'The providers on this phase have no credentials. Enter them on the Providers page before creating resource configurations.'}
        </div>
      )}

      {outlineConfigs.length > 0 && (
        <div className='rounded-lg border border-blue-300 bg-blue-50 p-3 text-sm text-blue-900'>
          <p className='font-medium'>How Outline access works</p>
          <p className='mt-1'>
            Keycloak signs the student in. Outline decides what they can see from its own group
            membership, not from the login token, because Outline has no group synchronisation. On
            each run PROMPT puts the same team members into the Keycloak group and into an Outline
            group, and grants that group access to the collection. Nothing is synchronised
            afterwards and no member is ever removed, so settle your teams before you provision.
          </p>
          {outlineWithoutKeycloak.length > 0 && (
            <p className='mt-2'>
              {outlineWithoutKeycloak.length === 1
                ? 'One Outline configuration has'
                : `${outlineWithoutKeycloak.length} Outline configurations have`}{' '}
              no Keycloak group configuration at the same scope. The collections will still be
              created and access granted through their Outline groups, but no matching Keycloak
              group is provisioned.
            </p>
          )}
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
