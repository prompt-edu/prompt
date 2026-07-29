import { test } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'lecturer' })

test.describe('application: student preview', () => {
  test('a lecturer previews the student-facing application form', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)

    await admin.gotoQuestions(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)
    await admin.openStudentPreview()
    await admin.expectPreviewLoaded()
  })
})
