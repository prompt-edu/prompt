import { ServiceStatusCard } from './components/ServiceStatusCard'
import { KNOWN_SERVICES } from './config/knownServices'

export const SystemStatusPage = () => {
  return (
    <div className='flex flex-col gap-6 w-full'>
      <h1 className='text-3xl font-bold tracking-tight'>System Status</h1>
      <div className='grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-3'>
        {KNOWN_SERVICES.map((service) => (
          <ServiceStatusCard key={service.apiBasePath} service={service} />
        ))}
      </div>
    </div>
  )
}
