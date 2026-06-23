import { useQuery } from '@tanstack/react-query'
import { axiosInstance } from '@tumaet/prompt-shared-state'
import { UserSearchResults } from '../interfaces/StaffMember'

const searchKeycloakUsers = async (query: string, limit = 20): Promise<UserSearchResults> => {
  return (
    await axiosInstance.get('/api/keycloak/users/search', {
      params: { q: query, limit },
    })
  ).data
}

export const useSearchKeycloakUsers = (query: string) => {
  const trimmed = query.trim()
  return useQuery({
    queryKey: ['keycloakUserSearch', trimmed],
    queryFn: () => searchKeycloakUsers(trimmed),
    enabled: trimmed.length >= 2,
    staleTime: 60_000,
  })
}
