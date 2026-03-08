import { useQueries } from '@tanstack/react-query'
import { ServiceStatusCard } from './components/ServiceStatusCard'
import { KNOWN_SERVICES } from './config/knownServices'
import { ServiceInfo } from './interfaces/serviceCapabilities'
import { getServiceInfo } from './network/getServiceCapabilities'

export const SystemStatusPage = () => {
  const results = useQueries({
    queries: KNOWN_SERVICES.map((service) => ({
      queryKey: ['serviceInfo', service.apiBasePath],
      queryFn: () => getServiceInfo(service),
      retry: false,
      staleTime: 30_000,
    })),
  })

  const availableServices = KNOWN_SERVICES.filter((_, i) => !results[i].isError)
  const unavailableServices = KNOWN_SERVICES.filter((_, i) => results[i].isError)

  const renderCard = (service: (typeof KNOWN_SERVICES)[number]) => {
    const i = KNOWN_SERVICES.indexOf(service)
    return (
      <ServiceStatusCard
        key={service.apiBasePath}
        service={service}
        data={results[i].data as ServiceInfo | undefined}
        isPending={results[i].isPending}
        isError={results[i].isError}
      />
    )
  }

  return (
    <div className='flex flex-col gap-8 w-full'>
      <h1 className='text-3xl font-bold tracking-tight'>System Status</h1>

      <div className='flex flex-col gap-3'>
        <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
          Available ({availableServices.length})
        </h2>
        {availableServices.length > 0 ? (
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {availableServices.map(renderCard)}
          </div>
        ) : (
          <p className='text-sm text-muted-foreground'>No services available</p>
        )}
      </div>

      {unavailableServices.length > 0 && (
        <div className='flex flex-col gap-3'>
          <h2 className='text-sm font-semibold uppercase tracking-wide text-muted-foreground'>
            Unavailable ({unavailableServices.length})
          </h2>
          <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
            {unavailableServices.map(renderCard)}
          </div>
        </div>
      )}
    </div>
  )
}
