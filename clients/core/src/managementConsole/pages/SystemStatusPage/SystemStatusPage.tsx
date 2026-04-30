import { useQueries } from '@tanstack/react-query'
import { ServiceStatusCard } from './components/ServiceStatusCard'
import { ServiceInfo } from './interfaces/serviceCapabilities'
import { CoursePhaseType } from './interfaces/coursePhaseType'
import { getServiceInfo } from './network/getServiceCapabilities'
import { useGetCoursePhaseTypes } from './hooks/useGetCoursePhaseTypes'

export const SystemStatusPage = () => {
  const { data: coursePhaseTypes = [] } = useGetCoursePhaseTypes()

  const results = useQueries({
    queries: coursePhaseTypes.map((service) => {
      return {
        queryKey: ['serviceInfo-' + service.id],
        queryFn: () => getServiceInfo(service),
        retry: false,
        staleTime: 30_000,
      }
    }),
  })

  const availableServices = coursePhaseTypes.filter((_, i) => !results[i].isError)
  const unavailableServices = coursePhaseTypes.filter((_, i) => results[i].isError)

  const renderCard = (service: CoursePhaseType) => {
    const i = coursePhaseTypes.indexOf(service)
    return (
      <ServiceStatusCard
        key={service.id}
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
