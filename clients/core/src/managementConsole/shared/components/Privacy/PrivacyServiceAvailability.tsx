import { useQueries, useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { Button, Popover, PopoverContent, PopoverTrigger } from '@tumaet/prompt-ui-components'
import { Loader2 } from 'lucide-react'
import type { ReactNode } from 'react'
import { useGetCoursePhaseTypes } from '../../../pages/SystemStatusPage/hooks/useGetCoursePhaseTypes'
import { getServiceInfo } from '../../../pages/SystemStatusPage/network/getServiceCapabilities'

enum PSAStatus {
  Loading = 'loading',
  Good = 'good',
  Problem = 'problem',
}

const GreenDot = <span className='h-2 w-2 rounded-full shrink-0 bg-green-500' />
const RedDot = <span className='h-2 w-2 rounded-full shrink-0 bg-destructive' />
const OrangeDot = <span className='h-2 w-2 rounded-full bg-orange-500 shrink-0' />

function ServiceList({
  services,
  indicator,
}: {
  services: { id: string; name: string }[]
  indicator: ReactNode
}) {
  return (
    <ul>
      {services.map((s) => (
        <li key={s.id} className='flex items-center gap-1.5 text-sm'>
          {indicator}
          {s.name}
        </li>
      ))}
    </ul>
  )
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
      {status === PSAStatus.Good && GreenDot}
      {status === PSAStatus.Problem && OrangeDot}
      <span>{message}</span>
    </div>
  )
}

interface PrivacyServiceAvailabilityProps {
  forSelf?: boolean
}

function hasOwnMicroservice(baseUrl: string): boolean {
  try {
    const parsed = new URL(baseUrl)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

export function PrivacyServiceAvailability({ forSelf }: PrivacyServiceAvailabilityProps = {}) {
  const { data: allCoursePhaseTypes = [], isPending: typesPending } =
    useGetCoursePhaseTypes(forSelf)
  const coursePhaseTypes = allCoursePhaseTypes.filter((cpt) => hasOwnMicroservice(cpt.baseUrl))

  const coreQuery = useQuery({
    queryKey: ['serviceInfo-core'],
    queryFn: () => axiosInstance.get('/api/hello'),
    retry: false,
    staleTime: 30_000,
  })

  const cpmResults = useQueries({
    queries: coursePhaseTypes.map((service) => ({
      queryKey: [`serviceInfo-${service.id}`],
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
        <Button variant='outline'>
          <PrivacyServiceAvailabilityIndicator status={status} message={message} />
        </Button>
      </PopoverTrigger>
      <PopoverContent align='start' className='w-auto p-3'>
        <p className='text-muted-foreground mb-2 max-w-64 text-center'>
          Services that are unavailable won&apos;t be able to respond to a deletion or export
          request
        </p>
        <hr className='my-1' />
        <div className='p-2'>
          <ServiceList services={allServices.filter((s) => !s.isError)} indicator={GreenDot} />
          <ServiceList services={allServices.filter((s) => s.isError)} indicator={RedDot} />
        </div>
      </PopoverContent>
    </Popover>
  )
}
