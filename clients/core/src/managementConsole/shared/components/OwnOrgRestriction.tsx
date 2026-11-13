import type { OrgRole } from '@core/interfaces/org'
import { useKeycloak } from '@core/keycloak/useKeycloak'
import { useOwnOrgs } from '@core/network/hooks/useOrgs'
import { useAuthStore } from '@tumaet/prompt-shared-state'
import { QueryGate, UnauthorizedPage } from '@tumaet/prompt-ui-components'
import type { ReactNode } from 'react'
import { useParams } from 'react-router-dom'
import { hasAnyOrgRole } from '../utils/orgRoles'

interface OwnOrgRestrictionProps {
  children: (role: OrgRole) => ReactNode
}

/**
 * Renders the page of the org in the URL for its Admins and Members, with the caller's role in it,
 * and the Unauthorized page for everyone else. The server's list of the caller's orgs decides, since
 * the token's roles name orgs by slug and the URL by id.
 */
export const OwnOrgRestriction = ({ children }: OwnOrgRestrictionProps) => {
  const { orgId } = useParams<{ orgId: string }>()
  const { permissions } = useAuthStore()
  const { logout } = useKeycloak()
  const holdsOrgRoles = hasAnyOrgRole(permissions)
  const ownOrgsQuery = useOwnOrgs({ enabled: holdsOrgRoles })

  if (!holdsOrgRoles) {
    return <UnauthorizedPage onLogout={logout} />
  }

  return (
    <QueryGate queries={[ownOrgsQuery]}>
      {() => {
        const ownOrg = ownOrgsQuery.data?.find((org) => org.id === orgId)
        return ownOrg ? children(ownOrg.role) : <UnauthorizedPage onLogout={logout} />
      }}
    </QueryGate>
  )
}
