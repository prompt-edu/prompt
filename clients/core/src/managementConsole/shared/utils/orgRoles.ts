import { type CourseOrg, ORG_ROLES, type OrgRole, type OwnOrg } from '@core/interfaces/org'
import { validateSlug } from '@core/managementConsole/orgAdministration/utils/validateOrgForm'
import { type Course, getPermissionString, Role } from '@tumaet/prompt-shared-state'

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

export const courseOrgID = (course: Course): string | undefined =>
  (course as Course & CourseOrg).orgID

/**
 * The Admins of an org are Lecturer-equivalent on all of its courses. The server grants that by
 * itself, but the console and the phase remotes check Lecturer rights by the course role in the
 * token. So the caller gets the course Lecturer role of every course of an org they are an Admin of.
 * These roles only open the UI: the server still decides every request.
 */
export const orgLecturerPermissions = (courses: Course[], ownOrgs: OwnOrg[]): string[] => {
  const adminOrgIDs = new Set(ownOrgs.filter((org) => org.role === 'Admin').map((org) => org.id))
  return courses
    .filter((course) => adminOrgIDs.has(courseOrgID(course) ?? ''))
    .map((course) => getPermissionString(Role.COURSE_LECTURER, course.name, course.semesterTag))
}
