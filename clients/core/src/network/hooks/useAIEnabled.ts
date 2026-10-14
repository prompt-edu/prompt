import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

// Not deployed, unreachable or unhealthy all hide the AI surfaces alike.
export const useAIStatus = (): { enabled: boolean; isPending: boolean } => {
  const info = useQuery({
    queryKey: coreKeys.ai.info(),
    queryFn: coreApi.ai.info,
    retry: false,
    staleTime: 60_000,
  })
  return {
    enabled: info.data?.serviceName === 'ai' && info.data.healthy === true,
    isPending: info.isPending,
  }
}

export const useAIEnabled = (): boolean => useAIStatus().enabled
