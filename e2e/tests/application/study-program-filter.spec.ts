import { test, expect } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { SEEDED_COURSES, FULL_COURSE_PHASES } from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.application.id

// The seeded fullCourse Application phase has participants across Computer
// Science and Information Systems (plus one with no study program), so the
// filter derives its options from real participant data.
test.use({ role: 'admin' })

test.describe('application: study program filter', () => {
  test('filters participants by study program', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)
    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, PHASE_ID)

    await admin.openFilterMenu()
    await expect(admin.studyProgramOption('Computer Science')).toBeVisible()
    await expect(admin.studyProgramOption('Information Systems')).toBeVisible()

    await admin.filterByStudyProgram('Information Systems')

    // An Information Systems applicant stays; a Computer Science one is filtered out.
    await expect(admin.applicantRow('test2@test.de')).toBeVisible()
    await expect(admin.applicantRow('niclas.heun@tum.de')).toHaveCount(0)
  })
})
