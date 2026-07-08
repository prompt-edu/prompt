import { test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// assessment_component REMOTE, loaded by the core shell from
// /assessment/remoteEntry.js through the e2e nginx proxy. If the remote fails
// to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('assessment: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.assessment.id)
    await phase.expectOverviewLoaded()
  })
})
