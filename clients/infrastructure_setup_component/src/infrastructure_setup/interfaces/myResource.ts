import type { ProviderType } from './providerConfig'
import type { Scope } from './resourceConfig'
import type { ResourceStatus } from './resourceInstance'

// One configured resource as a student of the phase sees it.
export interface MyResource {
  resourceConfigId: string
  providerType: ProviderType
  resourceType: string
  scope: Scope
  // Null when nothing has been provisioned for the student from this config yet.
  status: ResourceStatus | null
  // Null when there is no instance, or its run predates member tracking.
  granted: boolean | null
  name: string
  teamName: string
  // Null when there is nothing a student can open, such as a Keycloak group.
  url: string | null
}
