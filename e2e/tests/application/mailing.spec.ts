import { test } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'lecturer' })

test.describe('application: mailing configuration', () => {
  // Configuration UI only — no mail is sent from this page.
  test('the mailing settings page renders its configuration', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoMailing(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)
    await admin.expectMailingLoaded()
  })
})
