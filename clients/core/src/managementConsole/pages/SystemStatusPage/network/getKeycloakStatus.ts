import { axiosInstance } from '@tumaet/prompt-shared-state'
import { KeycloakStatus } from '../interfaces/keycloakStatus'

export const getKeycloakStatus = async (): Promise<KeycloakStatus> => {
  return (await axiosInstance.get<KeycloakStatus>('/api/keycloak/status')).data
}
