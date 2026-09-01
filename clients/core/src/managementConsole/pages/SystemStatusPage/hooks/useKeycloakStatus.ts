import { coreKeys } from '@core/network/cache'
import { useQuery } from '@tanstack/react-query'
import type { KeycloakStatus } from '../interfaces/keycloakStatus'
import { getKeycloakStatus } from '../network/getKeycloakStatus'

export function useKeycloakStatus() {
  return useQuery<KeycloakStatus>({
    queryKey: coreKeys.keycloak.status(),
    queryFn: getKeycloakStatus,
    retry: false,
    staleTime: 30_000,
  })
}
