import { test } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import {
  FULL_COURSE_APPLICATION_PARTICIPANTS,
  FULL_COURSE_PHASES,
  SEEDED_COURSES,
} from '../../src/data/constants'

const { stan, maxMustermann } = FULL_COURSE_APPLICATION_PARTICIPANTS

test.use({ role: 'lecturer' })

test.describe('application: participants list', () => {
  test('the participants page lists the seeded participations', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)

    await admin.expectParticipantListed(stan.email, stan.lastName)
    await admin.expectParticipantListed(maxMustermann.email, maxMustermann.lastName)
  })

  test('an accepted applicant has no hidden picture marker', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)

    await admin.expectPictureShownInList(stan.email)
    await admin.openApplication(stan.email)
    await admin.expectPictureShownOnApplication()
  })
})
