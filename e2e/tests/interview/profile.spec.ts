import type { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { test, expect } from '../../src/fixtures/auth'
import { InterviewPage } from '../../src/pages/InterviewPage'
import {
  SEEDED_COURSES,
  FULL_COURSE_PHASES,
  FULL_COURSE_APPLICATION_PARTICIPANTS,
} from '../../src/data/constants'

// The only uniquely-named interview participant (see lecturer-journey.spec.ts).
const STUDENT_NAME = 'Max Mustermann'
const STUDENT_ID = FULL_COURSE_APPLICATION_PARTICIPANTS.maxMustermann.id
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const COURSE_NAME = SEEDED_COURSES.fullCourse.name
const PHASE_ID = FULL_COURSE_PHASES.interview.id
const PROFILE_URL = `/management/course/${COURSE_ID}/${PHASE_ID}/manage/${STUDENT_ID}`

test.use({ role: 'course-lecturer' })

async function setShowStudentHistory(api: APIRequestContext, showStudentHistory: boolean) {
  const res = await api.put(`/api/course_phases/${PHASE_ID}`, {
    data: { id: PHASE_ID, restrictedData: { showStudentHistory } },
  })
  expect(res.ok(), await res.text()).toBeTruthy()
}

test.describe('interview: profile', () => {
  test('a lecturer opens a profile without a dead breadcrumb', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoOverview(COURSE_ID, PHASE_ID)
    await phase.expectOverviewLoaded()

    await phase.openProfile(STUDENT_NAME)

    await expect(page).toHaveURL(PROFILE_URL)
    await expect(phase.breadcrumb().getByText('Manage', { exact: true })).toBeVisible()
    await expect(phase.breadcrumb().getByText('Details', { exact: true })).toHaveCount(0)
  })

  test('an old details link redirects to the profile', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.goto(COURSE_ID, PHASE_ID, `/manage/details/${STUDENT_ID}`)

    await expect(page).toHaveURL(PROFILE_URL)
    await expect(page.getByText(STUDENT_NAME, { exact: true }).first()).toBeVisible()
  })

  test.describe('with student history enabled', () => {
    let api: APIRequestContext

    test.beforeAll(async () => {
      api = await apiContextFor('admin')
      await setShowStudentHistory(api, true)
    })

    test.afterAll(async () => {
      await setShowStudentHistory(api, false)
      await api.dispose()
    })

    test('the profile shows the course history loaded from core', async ({ page }) => {
      await page.goto(PROFILE_URL)

      await expect(
        page.getByRole('heading', { level: 3, name: COURSE_NAME, exact: true }),
      ).toBeVisible({ timeout: 15_000 })
    })
  })
})
