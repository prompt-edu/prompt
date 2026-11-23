import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'

export const useSearchKeycloakUsers = (query: string) => {
  const trimmed = query.trim()
  return useQuery({
    queryKey: coreKeys.keycloak.userSearch.forQuery(trimmed),
    queryFn: () => coreApi.keycloak.searchUsers(trimmed),
    enabled: trimmed.length >= 2,
    staleTime: 60_000,
  })
}
