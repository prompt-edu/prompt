import { test } from '../../src/fixtures/auth'
import { InterviewPage } from '../../src/pages/InterviewPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'

// Module Federation smoke test: the phase page is rendered by the
// interview_component REMOTE, loaded by the core shell from
// /interview/remoteEntry.js through the e2e nginx proxy. If the remote fails
// to load, the shell renders a LoadingError instead of the heading.
test.use({ role: 'course-lecturer' })

test.describe('interview: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.interview.id)
    await phase.expectLoaded()
  })
})
