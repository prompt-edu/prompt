import { test, authFile } from '../../src/fixtures/auth'
import { AuthenticatedApplicationPage } from '../../src/pages/AuthenticatedApplicationPage'
import {
  FULL_COURSE_APPLICATION_QUESTION,
  FULL_COURSE_PHASES,
  FULL_COURSE_STUDENT,
} from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.application.id
const EDITED_SEMESTER = '7'
const ORIGINAL_SEMESTER = String(FULL_COURSE_STUDENT.currentSemester)
const MOTIVATION_ANSWER = 'Keeping my profile up to date for the course.'

test.use({ role: 'student' })

test.describe('students: profile view and edit', () => {
  // Restore the seeded semester so parallel specs and reruns start from a known
  // profile; identity fields (name, matriculation) are never touched.
  test.afterAll(async ({ browser }) => {
    const context = await browser.newContext({ storageState: authFile('student') })
    try {
      const restore = new AuthenticatedApplicationPage(await context.newPage())
      await restore.goto(PHASE_ID)
      await restore.expectLoaded()
      await restore.saveProfile(ORIGINAL_SEMESTER, FULL_COURSE_APPLICATION_QUESTION.title, MOTIVATION_ANSWER)
    } finally {
      await context.close()
    }
  })

  test('a student views their prefilled profile and edits it', async ({ page }) => {
    const profile = new AuthenticatedApplicationPage(page)

    await profile.goto(PHASE_ID)
    await profile.expectLoaded()
    await profile.expectProfilePrefilled(
      FULL_COURSE_STUDENT.firstName,
      FULL_COURSE_STUDENT.matriculationNumber,
      ORIGINAL_SEMESTER,
    )

    await profile.saveProfile(EDITED_SEMESTER, FULL_COURSE_APPLICATION_QUESTION.title, MOTIVATION_ANSWER)

    await profile.goto(PHASE_ID)
    await profile.expectLoaded()
    await profile.expectProfilePrefilled(
      FULL_COURSE_STUDENT.firstName,
      FULL_COURSE_STUDENT.matriculationNumber,
      EDITED_SEMESTER,
    )
  })
})
