import { useQuery } from '@tanstack/react-query'
import { KeycloakStatus } from '../interfaces/keycloakStatus'
import { getKeycloakStatus } from '../network/getKeycloakStatus'

export function useKeycloakStatus() {
  return useQuery<KeycloakStatus>({
    queryKey: ['keycloakStatus'],
    queryFn: getKeycloakStatus,
    retry: false,
    staleTime: 30_000,
  })
}
