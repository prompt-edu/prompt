import { Popover, PopoverContent, PopoverTrigger } from '@tumaet/prompt-ui-components'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { useQueries, useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { getServiceInfo } from '../../../pages/SystemStatusPage/network/getServiceCapabilities'
import { useGetCoursePhaseTypes } from '../../../pages/SystemStatusPage/hooks/useGetCoursePhaseTypes'

enum PSAStatus {
  Loading = 'loading',
  Good = 'good',
  Problem = 'problem',
}

function PrivacyServiceAvailabilityIndicator({
  status,
  message,
}: {
  status: PSAStatus
  message: string
}) {
  return (
    <div className='flex items-center gap-1.5 text-sm text-muted-foreground'>
      {status === PSAStatus.Loading && <Loader2 className='h-3 w-3 animate-spin' />}
      {status === PSAStatus.Good && <span className='h-2 w-2 rounded-full bg-green-500 shrink-0' />}
      {status === PSAStatus.Problem && (
        <span className='h-2 w-2 rounded-full bg-orange-500 shrink-0' />
      )}
      <span className='text-black'>{message}</span>
    </div>
  )
}

export function PrivacyServiceAvailability() {
  const { data: coursePhaseTypes = [], isPending: typesPending } = useGetCoursePhaseTypes()

  const coreQuery = useQuery({
    queryKey: ['serviceInfo-core'],
    queryFn: () => axiosInstance.get('/api/hello'),
    retry: false,
    staleTime: 30_000,
  })

  const cpmResults = useQueries({
    queries: coursePhaseTypes.map((service) => ({
      queryKey: ['serviceInfo-' + service.id],
      queryFn: () => getServiceInfo(service),
      retry: false,
      staleTime: 30_000,
    })),
  })

  const isLoading = typesPending || coreQuery.isPending || cpmResults.some((r) => r.isPending)

  const allServices = [
    { id: 'core', name: 'Prompt Core', isError: coreQuery.isError },
    ...coursePhaseTypes.map((s, i) => ({ id: s.id, name: s.name, isError: cpmResults[i].isError })),
  ]

  const unavailableCount = allServices.filter((s) => s.isError).length

  const status: PSAStatus = (() => {
    if (isLoading) return PSAStatus.Loading
    if (unavailableCount === 0) return PSAStatus.Good
    return PSAStatus.Problem
  })()

  const message = (() => {
    switch (status) {
      case PSAStatus.Loading:
        return 'Checking availability...'
      case PSAStatus.Good:
        return 'All services available'
      case PSAStatus.Problem:
        return 'Some services are unavailable right now'
    }
  })()

  return (
    <Popover>
      <PopoverTrigger asChild>
        <button>
          <PrivacyServiceAvailabilityIndicator status={status} message={message} />
        </button>
      </PopoverTrigger>
      <PopoverContent align='start' className='w-auto p-3'>
        <ul className='space-y-1'>
          {allServices.map((s) => (
            <li key={s.id} className='flex items-center gap-1.5 text-sm'>
              <span
                className={`h-2 w-2 rounded-full shrink-0 ${s.isError ? 'bg-destructive' : 'bg-green-500'}`}
              />
              {s.name}
            </li>
          ))}
        </ul>
      </PopoverContent>
    </Popover>
  )
}
