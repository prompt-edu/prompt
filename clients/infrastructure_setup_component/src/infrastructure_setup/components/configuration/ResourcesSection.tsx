import { useQuery } from '@tanstack/react-query'
import { Button, Skeleton } from '@tumaet/prompt-ui-components'
import { PlusCircle } from 'lucide-react'
import { useState } from 'react'
import { ResourceConfigUpsertDialog } from '../../dialogs/ResourceConfigUpsertDialog'
import type { ResourceConfig } from '../../interfaces/resourceConfig'
import { getInstances } from '../../network/queries/getInstances'
import { getProviderConfigs } from '../../network/queries/getProviderConfigs'
import { getResourceConfigs } from '../../network/queries/getResourceConfigs'
import { ResourceConfigCard } from '../ResourceConfigCard'
import { SectionError } from '../SectionError'
import { ConfigurationSection } from './ConfigurationSection'

interface Props {
  coursePhaseID: string
}

export const ResourcesSection = ({ coursePhaseID }: Props) => {
  const [dialogOpen, setDialogOpen] = useState(false)
  const [editing, setEditing] = useState<ResourceConfig | undefined>(undefined)

  const {
    data: resourceConfigs,
    isLoading,
    isError,
    refetch,
  } = useQuery({
    queryKey: ['resource-configs', coursePhaseID],
    queryFn: () => getResourceConfigs(coursePhaseID),
  })

  const {
    data: providers,
    isLoading: providersLoading,
    isError: providersError,
    refetch: refetchProviders,
  } = useQuery({
    queryKey: ['provider-configs', coursePhaseID],
    queryFn: () => getProviderConfigs(coursePhaseID),
  })

  const {
    data: instances,
    isLoading: instancesLoading,
    isError: instancesError,
    refetch: refetchInstances,
  } = useQuery({
    queryKey: ['instances', coursePhaseID],
    queryFn: () => getInstances(coursePhaseID),
  })

  const loading = isLoading || providersLoading || instancesLoading
  // All three are needed before the section can say anything true: without the providers
  // the banner would claim none are configured and the create button would stay
  // disabled, and without the instances a config whose resource already exists would
  // offer edits the server refuses.
  const failed = isError || providersError || instancesError

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

  // A failed instance never created anything, so it does not pin the config's identity.
  const provisionedConfigIDs = new Set(
    (instances ?? []).filter((i) => i.status !== 'failed').map((i) => i.resourceConfigId),
  )

  const outlineConfigs = (resourceConfigs ?? []).filter((c) => c.providerType === 'outline')
  const keycloakScopes = new Set(
    (resourceConfigs ?? []).filter((c) => c.providerType === 'keycloak').map((c) => c.scope),
  )
  const outlineWithoutKeycloak = outlineConfigs.filter((c) => !keycloakScopes.has(c.scope))

  return (
    <ConfigurationSection
      id='resources'
      step={3}
      title='Resources'
      description='What to create for every team or every student, and who gets which permission. Teams come from the Team Allocation or Self Team Allocation phase connected to this one in the course configurator.'
      action={
        <Button
          onClick={openCreate}
          disabled={loading || failed || availableProviderTypes.length === 0}
        >
          <PlusCircle className='mr-2 h-4 w-4' />
          Add resource
        </Button>
      }
    >
      {loading ? (
        <Skeleton className='h-16 w-full' />
      ) : failed ? (
        <SectionError
          message='Failed to load the resources.'
          onRetry={() => {
            refetch()
            refetchProviders()
            refetchInstances()
          }}
        />
      ) : (
        <>
          {availableProviderTypes.length === 0 && (
            <div className='rounded-lg border border-amber-300 bg-amber-50 p-3 text-sm text-amber-900'>
              {(providers ?? []).length === 0
                ? 'Add a provider above before adding resources.'
                : 'The providers of this phase have no credentials. Enter them above before adding resources.'}
            </div>
          )}

          {outlineConfigs.length > 0 && (
            <div className='rounded-lg border border-blue-300 bg-blue-50 p-3 text-sm text-blue-900'>
              <p className='font-medium'>How Outline access works</p>
              <p className='mt-1'>
                Keycloak signs the student in. Outline decides what they can see from its own group
                membership, not from the login token, because Outline has no group synchronisation.
                On each run PROMPT puts the same team members into the Keycloak group and into an
                Outline group, and grants that group access to the collection. Nothing is
                synchronised afterwards and no member is ever removed, so settle your teams before
                you provision.
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
              No resource configured yet.
            </div>
          ) : (
            <div className='space-y-2'>
              {resourceConfigs.map((config) => (
                <ResourceConfigCard
                  key={config.id}
                  coursePhaseID={coursePhaseID}
                  config={config}
                  isProvisioned={provisionedConfigIDs.has(config.id)}
                  onEdit={openEdit}
                />
              ))}
            </div>
          )}
        </>
      )}

      <ResourceConfigUpsertDialog
        coursePhaseID={coursePhaseID}
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        existing={editing}
        identityLocked={!!editing && provisionedConfigIDs.has(editing.id)}
        availableProviderTypes={availableProviderTypes}
      />
    </ConfigurationSection>
  )
}
