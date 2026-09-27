import { test, expect } from '../../src/fixtures/auth'
import { InterviewPage } from '../../src/pages/InterviewPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'

// The only uniquely-named interview participant (see lecturer-journey.spec.ts).
const STUDENT_NAME = 'Max Mustermann'
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const PHASE_ID = FULL_COURSE_PHASES.interview.id

test.use({ role: 'course-lecturer' })

test.describe('interview: profile', () => {
  test('a lecturer opens a profile without a dead breadcrumb', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoOverview(COURSE_ID, PHASE_ID)
    await phase.expectOverviewLoaded()

    await phase.openProfile(STUDENT_NAME)

    await expect(phase.breadcrumb().getByText('Manage', { exact: true })).toBeVisible()
    await expect(phase.breadcrumb().getByText('Details', { exact: true })).toHaveCount(0)
  })

  test('an old details link redirects to the profile', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoOverview(COURSE_ID, PHASE_ID)
    await phase.expectOverviewLoaded()
    const studentId = await phase.openProfile(STUDENT_NAME)

    await phase.goto(COURSE_ID, PHASE_ID, `/manage/details/${studentId}`)

    await expect(page).toHaveURL(new RegExp(`/${PHASE_ID}/manage/${studentId}$`))
    await expect(page.getByText(STUDENT_NAME, { exact: true }).first()).toBeVisible()
  })
})
