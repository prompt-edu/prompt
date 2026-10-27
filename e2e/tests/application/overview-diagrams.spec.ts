import { test } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { expectChartsDrawn } from '../../src/pages/charts'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

test.use({ role: 'lecturer' })

test.describe('application: overview diagrams', () => {
  // The page loads every phase remote, so this also catches a remote whose
  // build provides the shared UI library with a recharts copy of its own.
  test('every overview diagram draws its chart', async ({ page }) => {
    const admin = new ApplicationAdminPage(page)
    await admin.gotoOverview(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)

    // Applications, Gender, Study Program and Semester Distribution.
    await expectChartsDrawn(page, 4)
  })
})
