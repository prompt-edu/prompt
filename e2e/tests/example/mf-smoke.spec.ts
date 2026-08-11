import { test } from '../../src/fixtures/auth'
import { ExamplePage } from '../../src/pages/ExamplePage'
import { SEEDED_COURSES, EXAMPLE_PHASE_ID } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// example_component REMOTE, loaded by the core shell from
// /example/remoteEntry.js through the e2e nginx proxy. If the remote fails to
// load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('example: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new ExamplePage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, EXAMPLE_PHASE_ID)
    await phase.expectLoaded()
  })
})
