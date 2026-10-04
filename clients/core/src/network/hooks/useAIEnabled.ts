import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

// Switched off, unreachable or unhealthy all hide the AI surfaces alike.
export const useAIStatus = (): { enabled: boolean; isPending: boolean } => {
  const status = useQuery({
    queryKey: coreKeys.ai.status(),
    queryFn: coreApi.ai.status,
    staleTime: Infinity,
  })
  const switchedOn = status.data?.enabled === true
  const info = useQuery({
    queryKey: coreKeys.ai.info(),
    queryFn: coreApi.ai.info,
    enabled: switchedOn,
    retry: false,
    staleTime: 60_000,
  })
  return {
    enabled: switchedOn && info.data?.serviceName === 'ai' && info.data.healthy === true,
    isPending: status.isPending || (switchedOn && info.isPending),
  }
}

export const useAIEnabled = (): boolean => useAIStatus().enabled
