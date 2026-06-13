import { useQuery } from '@tanstack/react-query'

interface UsePrivacyRequestFlowOptions<
  TLatest,
  TCreated extends { id: string },
  TStatus extends { status: string },
> {
  resource: string
  getLatest: () => Promise<TLatest>
  extractIdFromLatest: (latest: TLatest | undefined) => string | undefined
  createRequest: () => Promise<TCreated>
  getStatus: (id: string) => Promise<TStatus>
  isEndState: (status: TStatus['status']) => boolean
  pollIntervalMs?: number
}

export function usePrivacyRequestFlow<
  TLatest,
  TCreated extends { id: string },
  TStatus extends { status: string },
>({
  resource,
  getLatest,
  extractIdFromLatest,
  createRequest,
  getStatus,
  isEndState,
  pollIntervalMs = 3000,
}: UsePrivacyRequestFlowOptions<TLatest, TCreated, TStatus>) {
  const latestQuery = useQuery({
    queryKey: ['privacy', `${resource}-latest`],
    queryFn: getLatest,
  })

  const createQuery = useQuery({
    queryKey: ['privacy', `${resource}-create`],
    queryFn: createRequest,
    enabled: false,
  })

  const id = createQuery.data?.id ?? extractIdFromLatest(latestQuery.data)

  const statusQuery = useQuery({
    queryKey: ['privacy', `${resource}-status`, id],
    queryFn: () => getStatus(id!),
    enabled: !!id,
    refetchInterval: (query) => {
      const status = query.state.data?.status
      if (status && isEndState(status)) return false
      return pollIntervalMs
    },
  })

  const isPolling = !!id && (!statusQuery.data || !isEndState(statusQuery.data.status))

  return {
    latest: latestQuery.data,
    record: statusQuery.data,
    isPolling,
    isCreating: createQuery.isFetching,
    triggerCreate: () => createQuery.refetch(),
  }
}
