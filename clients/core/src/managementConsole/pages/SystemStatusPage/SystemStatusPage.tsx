import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQueries } from '@tanstack/react-query'
import { KeycloakStatusCard } from './components/KeycloakStatusCard'
import { ServiceStatusCard } from './components/ServiceStatusCard'
import { useGetCoursePhaseTypes } from './hooks/useGetCoursePhaseTypes'
import {
  buildPhaseStatusEntries,
  type PhaseStatusEntry,
  phaseCardStatus,
  type ServiceStatus,
} from './utils/phaseStatus'

export const SystemStatusPage = () => {
  const { data: allCoursePhaseTypes = [] } = useGetCoursePhaseTypes()
  const entries = buildPhaseStatusEntries(allCoursePhaseTypes, __PROMPT_REMOTES__)
  const serverEntries = entries.filter((entry) => entry.hasServerInfo)
  const clientEntries = entries.flatMap((entry) =>
    entry.remote ? [{ entry, remote: entry.remote }] : [],
  )

  const serverResults = useQueries({
    queries: serverEntries.map(({ phaseType }) => ({
      queryKey: coreKeys.serviceInfo.ofService(phaseType.id),
      queryFn: () => coreApi.system.serviceInfo(phaseType),
      retry: false,
      staleTime: 30_000,
    })),
  })

  const clientResults = useQueries({
    queries: clientEntries.map(({ remote }) => ({
      queryKey: coreKeys.clientInfo.ofRemote(remote.name),
      queryFn: () => coreApi.system.clientInfo(remote),
      retry: false,
      staleTime: 30_000,
    })),
  })

  const probesOf = (entry: PhaseStatusEntry) => ({
    server: serverResults[serverEntries.indexOf(entry)],
    client: clientResults[clientEntries.findIndex((clientEntry) => clientEntry.entry === entry)],
  })

  const entriesWithStatus = (status: ServiceStatus) =>
    entries.filter((entry) => {
      const { server, client } = probesOf(entry)
      return (phaseCardStatus(server, client) ?? 'Online') === status
    })
  const availableServices = entriesWithStatus('Online')
  const degradedServices = entriesWithStatus('OnlineUnhealthy')
  const unavailableServices = entriesWithStatus('Offline')

  const renderCard = (entry: PhaseStatusEntry) => (
    <ServiceStatusCard key={entry.phaseType.id} entry={entry} {...probesOf(entry)} />
  )

  return (
    <div className='flex flex-col gap-8 w-full'>
      <h1 className='text-3xl font-bold tracking-tight'>System Status</h1>

      <div className='flex flex-col gap-3'>
        <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
          Infrastructure
        </h2>
        <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
          <KeycloakStatusCard />
        </div>
      </div>

      <div className='flex flex-col gap-3'>
        <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
          Available
        </h2>
        {availableServices.length > 0 ? (
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {availableServices.map(renderCard)}
          </div>
        ) : (
          <p className='text-sm text-muted-foreground'>None</p>
        )}
      </div>

      {degradedServices.length > 0 && (
        <div className='flex flex-col gap-3'>
          <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
            Degraded
          </h2>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {degradedServices.map(renderCard)}
          </div>
        </div>
      )}

      {unavailableServices.length > 0 && (
        <div className='flex flex-col gap-3'>
          <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
            Unavailable
          </h2>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {unavailableServices.map(renderCard)}
          </div>
        </div>
      )}
    </div>
  )
}
