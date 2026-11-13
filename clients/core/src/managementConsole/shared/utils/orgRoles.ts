import { ORG_ROLES, type OrgRole } from '@core/interfaces/org'
import { validateSlug } from '@core/managementConsole/orgAdministration/utils/validateOrgForm'

export interface ParsedOrgRole {
  slug: string
  role: OrgRole
}

const ORG_ROLE_PREFIX = 'org-'

/**
 * Mirrors `ParseOrgRole` on the server: accepts `org-<slug>-Admin` and `org-<slug>-Member` with a
 * valid org slug, and nothing else. A prefix and suffix match is not enough, because the custom-group
 * role `org-<course>-cg-Admin` of a course in semester `org` has both; the slug rules forbid the `cg`
 * segment that tells the two apart. Slugs may contain hyphens, so the role is never split on `-`.
 */
export const parseOrgRole = (permission: string): ParsedOrgRole | undefined => {
  if (!permission.startsWith(ORG_ROLE_PREFIX)) {
    return undefined
  }
  const rest = permission.slice(ORG_ROLE_PREFIX.length)
  for (const role of ORG_ROLES) {
    const suffix = `-${role}`
    const slug = rest.slice(0, -suffix.length)
    if (rest.endsWith(suffix) && validateSlug(slug) === undefined) {
      return { slug, role }
    }
  }
  return undefined
}

export const hasAnyOrgRole = (permissions: string[]): boolean =>
  permissions.some((permission) => parseOrgRole(permission) !== undefined)

// The org roles among the token's roles, sorted, so the same roles always give the same list
export const orgRolesOf = (permissions: string[]): string[] =>
  permissions.filter((permission) => parseOrgRole(permission) !== undefined).sort()
