import { PRESENTATION_PHASE_ID, SEEDED_COURSES } from '../../src/data/constants'
import { test } from '../../src/fixtures/auth'
import { PresentationPage } from '../../src/pages/PresentationPage'

// Verifies the complete same-origin integration: the core shell loads
// presentation_component from /presentation/remoteEntry.js and the remote
// fetches its overview through /presentation/api.
test.use({ role: 'course-lecturer' })

test.describe('presentation: module federation smoke', () => {
  test('the remote and phase API load inside the core shell', async ({ page }) => {
    const phase = new PresentationPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, PRESENTATION_PHASE_ID)
    await phase.expectOverviewLoaded()
  })
})
