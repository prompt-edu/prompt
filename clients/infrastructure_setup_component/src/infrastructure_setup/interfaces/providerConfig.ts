export type ProviderType = 'gitlab' | 'slack' | 'outline' | 'rancher' | 'keycloak'

export const providerTypes: ProviderType[] = ['gitlab', 'slack', 'outline', 'rancher', 'keycloak']

export interface ProviderConfig {
  id: string
  providerType: ProviderType
}

export interface AuthField {
  name: string
  label: string
  type: 'text' | 'password'
  required: boolean
  description: string
}

export interface UpsertProviderConfigRequest {
  providerType: ProviderType
  credentials: Record<string, unknown>
}
