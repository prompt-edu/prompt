import { test } from '../../src/fixtures/auth'
import { AssessmentPage } from '../../src/pages/AssessmentPage'
import { expectChartsDrawn } from '../../src/pages/charts'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'course-lecturer' })

test.describe('assessment: participants diagrams', () => {
  // The charts are composed inside the assessment remote while their container
  // comes from the shared UI library, so both must resolve one recharts module.
  test('every participants diagram draws its chart', async ({ page }) => {
    const phase = new AssessmentPage(page)
    await phase.goto(
      SEEDED_COURSES.fullCourse.id,
      FULL_COURSE_PHASES.assessment.id,
      '/participants',
    )
    await phase.expectParticipantsLoaded()

    // Assessment status, grade distribution and score level distribution. Each
    // draws a mark even before anyone is graded, so the lecturer journey grading
    // this phase in parallel cannot change the outcome.
    await expectChartsDrawn(page, 3)
  })
})
