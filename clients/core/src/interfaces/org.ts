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
