import type { UpdateOrg } from '@core/interfaces/org'
import { describe, expect, it } from 'vitest'
import { rebaseOrgDetails } from './rebaseOrgDetails'

const SAVED: UpdateOrg = {
  name: 'Demo Chair',
  school: 'CIT',
  university: 'TUM',
  website: 'https://aet.cit.tum.de',
  contactEmail: 'chair@example.com',
}

describe('rebaseOrgDetails', () => {
  it('takes the newer saved values of a draft without edits', () => {
    const after = { ...SAVED, name: 'Renamed Chair' }

    expect(rebaseOrgDetails(SAVED, after, SAVED)).toEqual({ draft: after, conflicts: [] })
  })

  it('keeps the edits and takes the newer values of the other fields', () => {
    const draft = { ...SAVED, contactEmail: 'new@example.com' }
    const after = { ...SAVED, name: 'Renamed Chair' }

    expect(rebaseOrgDetails(SAVED, after, draft)).toEqual({
      draft: { ...SAVED, name: 'Renamed Chair', contactEmail: 'new@example.com' },
      conflicts: [],
    })
  })

  it('reports a field both sides changed to different values', () => {
    const draft = { ...SAVED, name: 'My Name' }
    const after = { ...SAVED, name: 'Their Name' }

    expect(rebaseOrgDetails(SAVED, after, draft)).toEqual({ draft, conflicts: ['name'] })
  })

  it('reports no conflict when the saved value caught up with the edit', () => {
    // The refetch after a save brings back the saved edit
    const draft = { ...SAVED, school: 'School of CIT' }
    const after = { ...SAVED, school: 'School of CIT' }

    expect(rebaseOrgDetails(SAVED, after, draft)).toEqual({ draft, conflicts: [] })
  })

  it('keeps an edit made after a save while its refetch is still loading', () => {
    const saved = { ...SAVED, school: 'School of CIT' }
    const draft = { ...saved, university: 'Technical University of Munich' }

    expect(rebaseOrgDetails(SAVED, saved, draft)).toEqual({ draft, conflicts: [] })
  })
})
