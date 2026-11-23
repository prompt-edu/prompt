import type { OwnOrg } from '@core/interfaces/org'
import { useOwnOrgs } from '@core/network/hooks/useOrgs'
import { type Course, Role, useAuthStore } from '@tumaet/prompt-shared-state'
import { courseOrgID } from '../utils/orgRoles'

export interface NewCourseOrg {
  orgs: OwnOrg[]
  required: boolean
}

/**
 * The orgs a new course can go into: the caller's own. Only PROMPT admins and lecturers create
 * ungrouped courses, so everyone else with an org picks one. That includes copying a course of an
 * org, the source, which the caller may only copy through their org role.
 */
export const useNewCourseOrg = (source?: Course): NewCourseOrg => {
  const { permissions } = useAuthStore()
  const { data: orgs = [] } = useOwnOrgs()
  const hasGlobalRole =
    permissions.includes(Role.PROMPT_ADMIN) || permissions.includes(Role.PROMPT_LECTURER)
  const required =
    !hasGlobalRole && orgs.length > 0 && (source === undefined || courseOrgID(source) !== undefined)
  return { orgs, required }
}
