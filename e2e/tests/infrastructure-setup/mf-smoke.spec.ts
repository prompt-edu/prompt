import { test } from '../../src/fixtures/auth'
import { InfrastructureSetupPage } from '../../src/pages/InfrastructureSetupPage'
import { SEEDED_COURSES, INFRASTRUCTURE_SETUP_PHASE_ID } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// infrastructure_setup_component REMOTE, loaded by the core shell from
// /infrastructure-setup/remoteEntry.js through the e2e nginx proxy. If the remote
// fails to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('infrastructure setup: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new InfrastructureSetupPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, INFRASTRUCTURE_SETUP_PHASE_ID)
    await phase.expectLoaded()
  })
})
