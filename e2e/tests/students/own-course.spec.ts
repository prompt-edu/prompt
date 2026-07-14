import { test } from '../../src/fixtures/auth'
import { CourseOverviewPage } from '../../src/pages/CourseOverviewPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'student' })

test.describe('students: own course view', () => {
  test('a student sees their enrolled course and its phases', async ({ page }) => {
    const overview = new CourseOverviewPage(page)

    await overview.goto(SEEDED_COURSES.fullCourse.id)
    await overview.expectLoaded(SEEDED_COURSES.fullCourse.name)
    await overview.expectPhaseListed(FULL_COURSE_PHASES.interview.type)
    await overview.expectPhaseListed(FULL_COURSE_PHASES.assessment.type)
  })
})
