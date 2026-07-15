import { test } from '../../src/fixtures/auth'
import { MatchingPage } from '../../src/pages/MatchingPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'student' })

test.describe('matching: student access', () => {
  test('an enrolled student cannot see the lecturer-only matching phase', async ({ page }) => {
    const phase = new MatchingPage(page)
    await phase.goto(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.matching.id)
    await phase.expectAccessDenied()
  })
})
