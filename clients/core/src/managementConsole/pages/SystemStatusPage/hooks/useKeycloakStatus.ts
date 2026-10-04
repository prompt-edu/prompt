import { coreApi } from '@core/network/api'
import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'
import type { KeycloakStatus } from '../interfaces/keycloakStatus'

export function useKeycloakStatus() {
  return useQuery<KeycloakStatus>({
    queryKey: coreKeys.keycloak.status(),
    queryFn: coreApi.keycloak.status,
    retry: false,
    staleTime: 30_000,
  })
}
