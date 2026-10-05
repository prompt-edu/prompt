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
const STUDENT = FULL_COURSE_APPLICATION_PARTICIPANTS.maxMustermann
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const COURSE_NAME = SEEDED_COURSES.fullCourse.name
const PHASE_ID = FULL_COURSE_PHASES.interview.id
const PARTICIPANTS_URL = `/management/course/${COURSE_ID}/${PHASE_ID}/participants`
const PROFILE_URL = `${PARTICIPANTS_URL}/${STUDENT.id}`

test.use({ role: 'course-lecturer' })

async function setShowStudentHistory(api: APIRequestContext, showStudentHistory: boolean) {
  const res = await api.put(`/api/course_phases/${PHASE_ID}`, {
    data: { id: PHASE_ID, restrictedData: { showStudentHistory } },
  })
  expect(res.ok(), await res.text()).toBeTruthy()
}

test.describe('interview: profile', () => {
  test('a lecturer opens a profile from the participants table', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoParticipants(COURSE_ID, PHASE_ID)
    await phase.expectParticipantsLoaded()

    await phase.openProfile(STUDENT.firstName, STUDENT.lastName)

    await expect(page).toHaveURL(PROFILE_URL)
    await expect(phase.breadcrumb().getByText('Participants', { exact: true })).toBeVisible()
    await expect(phase.breadcrumb().getByText('Details', { exact: true })).toHaveCount(0)
  })

  test('Back from a profile keeps the table search', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoParticipants(COURSE_ID, PHASE_ID)
    await phase.expectParticipantsLoaded()
    await phase.searchParticipants(STUDENT.lastName)
    await expect(page).toHaveURL(`${PARTICIPANTS_URL}?search=${STUDENT.lastName}`)

    await phase.openProfile(STUDENT.firstName, STUDENT.lastName)
    await expect(page).toHaveURL(PROFILE_URL)
    await phase.backToParticipants()

    await expect(page).toHaveURL(`${PARTICIPANTS_URL}?search=${STUDENT.lastName}`)
    await expect(
      page.getByRole('row', { name: new RegExp(`${STUDENT.firstName} ${STUDENT.lastName}`) }),
    ).toBeVisible()
  })

  for (const { oldPath, newUrl } of [
    { oldPath: '/manage', newUrl: PARTICIPANTS_URL },
    { oldPath: `/manage/${STUDENT.id}`, newUrl: PROFILE_URL },
    { oldPath: `/manage/details/${STUDENT.id}`, newUrl: PROFILE_URL },
  ]) {
    test(`an old ${oldPath.replace(STUDENT.id, ':studentId')} link redirects`, async ({ page }) => {
      const phase = new InterviewPage(page)
      await phase.goto(COURSE_ID, PHASE_ID, oldPath)

      await expect(page).toHaveURL(newUrl)
    })
  }

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
