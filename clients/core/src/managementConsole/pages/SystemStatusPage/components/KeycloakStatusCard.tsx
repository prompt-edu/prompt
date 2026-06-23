import {
  Button,
  Card,
  CardContent,
  CardHeader,
  CardTitle,
  Separator,
  Skeleton,
} from '@tumaet/prompt-ui-components'
import { CheckCircle2, KeyRound, RefreshCw, XCircle } from 'lucide-react'
import { useKeycloakStatus } from '../hooks/useKeycloakStatus'
import { KEYCLOAK_CAPABILITY_LABELS } from '../interfaces/keycloakStatus'
import { ServiceStatusBadge } from './ServiceStatusBadge'
import { ServiceStatusCardSkeleton } from './ServiceStatusCardSkeleton'

const ALL_KEYCLOAK_CAPABILITIES = Object.keys(KEYCLOAK_CAPABILITY_LABELS)

export const KeycloakStatusCard = () => {
  const { data, isPending, isError, isFetching, refetch } = useKeycloakStatus()

  const capabilities = data?.capabilities ?? {}
  const isHealthy = data?.healthy === true
  const status = isError || data == null ? 'Offline' : !isHealthy ? 'OnlineUnhealthy' : 'Online'

  return (
    <Card>
      <CardHeader className='pb-3'>
        <div className='flex items-center justify-between'>
          <div className='flex items-center gap-2'>
            <KeyRound className='h-4 w-4 text-muted-foreground' />
            <CardTitle className='text-base'>Keycloak</CardTitle>
          </div>
          <div className='flex items-center gap-2'>
            {isPending ? (
              <Skeleton className='h-5 w-16' />
            ) : (
              <ServiceStatusBadge status={status} />
            )}
            <Button
              variant='ghost'
              size='icon'
              aria-label='Refresh Keycloak status'
              disabled={isFetching}
              onClick={() => refetch()}
            >
              <RefreshCw className={`h-4 w-4 ${isFetching ? 'animate-spin' : ''}`} />
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className='flex flex-col gap-4'>
        {isPending ? (
          <ServiceStatusCardSkeleton />
        ) : isError ? (
          <p className='text-sm text-muted-foreground'>
            Status endpoint unreachable - the core server may be offline or the route is gated.
          </p>
        ) : (
          <>
            <dl className='grid grid-cols-[auto_1fr] gap-x-4 gap-y-1 text-sm'>
              <dt className='text-muted-foreground'>Service account</dt>
              <dd className={isHealthy ? 'text-green-600' : 'text-yellow-600'}>
                {isHealthy ? 'Configured correctly' : 'Misconfigured or missing permission'}
              </dd>
            </dl>
            <Separator />
            <div className='flex flex-col gap-1.5'>
              <p className='text-xs font-semibold uppercase tracking-wide text-muted-foreground'>
                Checks
              </p>
              <ul className='flex flex-col gap-1.5'>
                {ALL_KEYCLOAK_CAPABILITIES.map((key) => {
                  const passed = capabilities[key] === true
                  return (
                    <li key={key} className='flex items-center gap-2 text-sm'>
                      {passed ? (
                        <CheckCircle2 className='h-4 w-4 shrink-0 text-green-500' />
                      ) : (
                        <XCircle className='h-4 w-4 shrink-0 text-red-500' />
                      )}
                      <span className={passed ? '' : 'text-muted-foreground'}>
                        {KEYCLOAK_CAPABILITY_LABELS[key]}
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
