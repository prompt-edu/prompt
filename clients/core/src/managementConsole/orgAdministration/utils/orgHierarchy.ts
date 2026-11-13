import type { Org } from '@core/interfaces/org'

export interface OrgRow extends Org {
  parentLabel: string | null
}

export const orgLabel = (org: Pick<Org, 'name' | 'slug'>): string => `${org.name} (${org.slug})`

export const childOrgs = (orgs: Org[], orgId: string): Org[] =>
  orgs.filter((org) => org.parentOrgID === orgId)

export const descendantIds = (orgs: Org[], orgId: string): Set<string> => {
  const childrenByParent = new Map<string, string[]>()
  for (const org of orgs) {
    if (org.parentOrgID !== null) {
      childrenByParent.set(org.parentOrgID, [
        ...(childrenByParent.get(org.parentOrgID) ?? []),
        org.id,
      ])
    }
  }

  const descendants = new Set<string>()
  const pending = [orgId]
  while (pending.length > 0) {
    for (const childId of childrenByParent.get(pending.pop() ?? '') ?? []) {
      if (!descendants.has(childId)) {
        descendants.add(childId)
        pending.push(childId)
      }
    }
  }
  return descendants
}

export const parentCandidates = (orgs: Org[], orgId?: string): Org[] => {
  if (orgId === undefined) {
    return orgs
  }
  const excluded = descendantIds(orgs, orgId).add(orgId)
  return orgs.filter((org) => !excluded.has(org.id))
}

export const withParentLabels = (orgs: Org[]): OrgRow[] => {
  const orgsById = new Map(orgs.map((org) => [org.id, org]))
  return orgs.map((org) => {
    const parent = org.parentOrgID === null ? undefined : orgsById.get(org.parentOrgID)
    return { ...org, parentLabel: parent ? orgLabel(parent) : null }
  })
}
