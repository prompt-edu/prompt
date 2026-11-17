import type { UpdateOrg } from '@core/interfaces/org'

export type OrgDetailsField = keyof UpdateOrg

export interface RebasedOrgDetails {
  draft: UpdateOrg
  // Fields both the draft and the saved org changed, to different values
  conflicts: OrgDetailsField[]
}

const FIELDS: OrgDetailsField[] = ['name', 'school', 'university', 'website', 'contactEmail']

// Moves a draft from the saved details it started from onto newer saved details: a field the draft
// left alone takes the newer value, and a field it edited keeps the edit.
export const rebaseOrgDetails = (
  before: UpdateOrg,
  after: UpdateOrg,
  draft: UpdateOrg,
): RebasedOrgDetails => {
  const rebased = { ...draft }
  const conflicts: OrgDetailsField[] = []
  for (const field of FIELDS) {
    if (draft[field] === before[field]) {
      rebased[field] = after[field]
    } else if (after[field] !== before[field] && after[field] !== draft[field]) {
      conflicts.push(field)
    }
  }
  return { draft: rebased, conflicts }
}
