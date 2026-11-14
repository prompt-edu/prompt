export interface Org {
  id: string
  parentOrgID: string | null
  name: string
  slug: string
  school: string | null
  university: string | null
  website: string | null
  contactEmail: string | null
  createdAt: string
}

export interface UpdateOrg {
  name: string
  school: string
  university: string
  website: string
  contactEmail: string
}

export interface CreateOrg extends UpdateOrg {
  slug: string
  parentOrgID: string | null
}

export interface UpdateOrgParent {
  parentOrgID: string | null
}

// The Keycloak role suffixes; the display labels may differ (#1789, question 1)
export type OrgRole = 'Admin' | 'Member'

export const ORG_ROLES: OrgRole[] = ['Admin', 'Member']

export const ORG_ROLE_LABELS: Record<OrgRole, string> = {
  Admin: 'Admin',
  Member: 'Member',
}

export interface OwnOrg extends Org {
  role: OrgRole
}

export interface OrgMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
  universityLogin: string
  role: OrgRole
}

export interface SetOrgMember {
  universityLogin: string
  role: OrgRole
}

export interface UpdateOrgMemberRole {
  role: OrgRole
}
