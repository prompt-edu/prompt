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
const PHASE_ID = FULL_COURSE_PHASES.interview.id
const PARTICIPANTS_URL = `/management/course/${COURSE_ID}/${PHASE_ID}/participants`
const PROFILE_URL = `${PARTICIPANTS_URL}/${STUDENT.id}`

test.use({ role: 'course-lecturer' })

test.describe('interview: profile', () => {
  test('a lecturer opens a profile from the participants table', async ({ page }) => {
    const phase = new InterviewPage(page)
    await phase.gotoParticipants(COURSE_ID, PHASE_ID)
    await phase.expectParticipantsLoaded()

    await phase.openProfile(STUDENT.firstName, STUDENT.lastName)

    await expect(page).toHaveURL(PROFILE_URL)
    await expect(phase.breadcrumb().getByText('Participants', { exact: true })).toBeVisible()
    await expect(
      phase.breadcrumb().getByText(`${STUDENT.firstName} ${STUDENT.lastName}`, { exact: true }),
    ).toBeVisible()
    await expect(phase.breadcrumb().getByText('Details', { exact: true })).toHaveCount(0)
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
})
