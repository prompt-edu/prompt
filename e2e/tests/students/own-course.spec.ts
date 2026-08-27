import { test } from '../../src/fixtures/auth'
import { CourseOverviewPage } from '../../src/pages/CourseOverviewPage'
import { TeamAllocationPage } from '../../src/pages/TeamAllocationPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'student' })

test.describe('students: own course view', () => {
  test('a student sees their enrolled course and its phases', async ({ page }) => {
    const overview = new CourseOverviewPage(page)

    await overview.goto(SEEDED_COURSES.fullCourse.id)
    await overview.expectLoaded(SEEDED_COURSES.fullCourse.name)
    await overview.expectPhaseListed(FULL_COURSE_PHASES.interview.type)
    await overview.expectPhaseListed(FULL_COURSE_PHASES.teamAllocation.type)
    await overview.expectPhaseListed(FULL_COURSE_PHASES.assessment.type)
  })

  test('a student opens a phase from the sidebar', async ({ page }) => {
    const overview = new CourseOverviewPage(page)

    await overview.goto(SEEDED_COURSES.fullCourse.id)
    await overview.expectLoaded(SEEDED_COURSES.fullCourse.name)
    await overview.openPhase(FULL_COURSE_PHASES.teamAllocation.type)

    await new TeamAllocationPage(page).expectSurveyLoaded()
  })
})

// `student2` (Selma) is enrolled in iPraktikumFull but active in none of its graph
// phases, and holds no course-scoped Keycloak role for it, so no phase sidebar
// entry may show: a top-level entry without `requiredPermissions` used to render
// for every course member regardless of phase membership.
test.describe('students: a course whose phases they are not in', () => {
  test.use({ role: 'student2' })

  test('a student does not see a phase they are not part of', async ({ page }) => {
    const overview = new CourseOverviewPage(page)

    await overview.goto(SEEDED_COURSES.fullCourse.id)
    await overview.expectLoaded(SEEDED_COURSES.fullCourse.name)
    await overview.expectNoPhaseListed(FULL_COURSE_PHASES.teamAllocation.type)
  })
})
