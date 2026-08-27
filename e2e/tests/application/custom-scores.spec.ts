import { test } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { CUSTOM_SCORES_APPLICATION, SEEDED_COURSES } from '../../src/data/constants'

const { phaseId, scoreName, scoredApplicant, unscoredApplicant } = CUSTOM_SCORES_APPLICATION

test.use({ role: 'lecturer' })

test.describe('application: custom scores', () => {
  test('the details page shows the uploaded score of an applicant', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, phaseId)
    await admin.openApplication(scoredApplicant.email)

    await admin.expectCustomScore(scoreName, scoredApplicant.score)
  })

  test('the details page marks an applicant without an uploaded score', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, phaseId)
    await admin.openApplication(unscoredApplicant.email)

    await admin.expectCustomScoreMissing(scoreName)
  })
})
