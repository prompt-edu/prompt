import { applications } from './applications'
import { apply } from './apply'
import { auditLog } from './auditLog'
import { courseGraphs } from './courseGraphs'
import { coursePhases } from './coursePhases'
import { courses } from './courses'
import { instructorNotes } from './instructorNotes'
import { keycloak } from './keycloak'
import { mailCampaigns } from './mailCampaigns'
import { mailing } from './mailing'
import { privacy } from './privacy'
import { students } from './students'
import { system } from './system'

/**
 * Every endpoint core calls, grouped by the domain noun its route already uses, so a call reads
 * `coreApi.applications.form(phaseId)` and the resource id comes first.
 */
export const coreApi = {
  applications,
  apply,
  auditLog,
  courseGraphs,
  coursePhases,
  courses,
  instructorNotes,
  keycloak,
  mailCampaigns,
  mailing,
  privacy,
  students,
  system,
}

export { buildAuditLogQuery } from './auditLog'
export type { SendStatusMailRequest } from './mailing'
