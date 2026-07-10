import { test } from '../../src/fixtures/auth'
import { SelfTeamAllocationPage } from '../../src/pages/SelfTeamAllocationPage'
import { SEEDED_COURSES, SELF_TEAM_ALLOCATION_PHASE_ID } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// self_team_allocation_component REMOTE, loaded by the core shell from
// /self-team-allocation/remoteEntry.js through the e2e nginx proxy. If the
// remote fails to load, the shell renders a LoadingError instead of the
// heading. This is the copyable per-module smoke test.
test.use({ role: 'course-lecturer' })

test.describe('self team allocation: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new SelfTeamAllocationPage(page)
    await phase.goto(SEEDED_COURSES.iPraktikum.id, SELF_TEAM_ALLOCATION_PHASE_ID)
    await phase.expectLoaded()
  })
})
