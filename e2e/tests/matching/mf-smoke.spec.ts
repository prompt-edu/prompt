import { test } from '../../src/fixtures/auth'
import { MatchingPage } from '../../src/pages/MatchingPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'course-lecturer' })

test.describe('matching: module federation smoke', () => {
  test('the remote loads and renders inside the core shell', async ({ page }) => {
    const phase = new MatchingPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.matching.id)
    await phase.expectOverviewLoaded()
  })
})
