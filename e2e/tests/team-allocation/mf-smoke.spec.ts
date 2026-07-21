import { test } from '../../src/fixtures/auth'
import { TeamAllocationPage } from '../../src/pages/TeamAllocationPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

// Module Federation smoke test: the /allocations view is rendered by the
// team_allocation_component REMOTE, loaded by the core shell from
// /team-allocation/remoteEntry.js through the e2e nginx proxy. If the remote
// fails to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('team allocation: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new TeamAllocationPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.teamAllocation.id, '/allocations')
    await phase.expectAllocationsLoaded()
  })
})
