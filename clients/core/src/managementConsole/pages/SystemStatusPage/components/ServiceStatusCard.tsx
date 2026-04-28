import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Separator,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { CheckCircle2, Server, XCircle } from 'lucide-react'
import { CAPABILITY_LABELS, ServiceInfo } from '../interfaces/serviceCapabilities'
import { CoursePhaseType } from '../interfaces/coursePhaseType'
import { ServiceStatusBadge } from './ServiceStatusBadge'
import { ServiceStatusCardSkeleton } from './ServiceStatusCardSkeleton'

interface ServiceStatusCardProps {
  service: CoursePhaseType
  data: ServiceInfo | undefined
  isPending: boolean
  isError: boolean
}

const ALL_KNOWN_CAPABILITIES = Object.keys(CAPABILITY_LABELS)

export const ServiceStatusCard = ({
  service,
  data,
  isPending,
  isError,
}: ServiceStatusCardProps) => {
  const capabilities = data?.capabilities ?? {}
  const isHealthy = data?.healthy === true

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <Server className='h-4 w-4 text-muted-foreground' />
            <CardTitle className='text-base'>{service.name}</CardTitle>
          </div>
          {isPending ? (
            <Skeleton className='h-5 w-16' />
          ) : (
            <ServiceStatusBadge
              status={(() => {
                if (isError || data == null) {
                  return 'Offline'
                }
                if (!isHealthy) {
                  return 'OnlineUnhealthy'
                } else {
                  return 'Online'
                }
              })()}
            />
          )}
        </div>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        {isPending ? (
          <ServiceStatusCardSkeleton />
        ) : isError ? (
          <p className='text-sm text-muted-foreground'>Service unreachable</p>
        ) : (
          <>
            <dl className='grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm'>
              <dt className='text-muted-foreground'>Version</dt>
              <dd>{data?.version ? `${data.version}` : 'n/a'}</dd>
              <dt className='text-muted-foreground'>Health</dt>
              <dd className={isHealthy ? 'text-green-600' : 'text-yellow-600'}>
                {isHealthy ? 'Healthy' : 'Degraded'}
              </dd>
            </dl>
            <Separator />
            <div className='flex flex-col gap-1.5'>
              <p className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'>
                Capabilities
              </p>
              <ul className='flex flex-col gap-1.5'>
                {ALL_KNOWN_CAPABILITIES.map((key) => {
                  const supported = capabilities[key] === true
                  return (
                    <li key={key} className='flex items-center gap-2 text-sm'>
                      {supported ? (
                        <CheckCircle2 className='h-4 w-4 shrink-0 text-green-500' />
                      ) : (
                        <XCircle className='h-4 w-4 shrink-0 text-muted-foreground' />
                      )}
                      <span className={supported ? '' : 'text-muted-foreground'}>
                        {CAPABILITY_LABELS[key]}
                      </span>
                    </li>
                  )
                })}
              </ul>
            </div>
          </>
        )}
      </CardContent>
    </Card>
  )
}
