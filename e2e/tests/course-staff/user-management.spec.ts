import { test } from '../../src/fixtures/auth'
import { CourseUserManagementPage } from '../../src/pages/CourseUserManagementPage'
import { SEEDED_COURSES } from '../../src/data/constants'

test.describe('course user management access control', () => {
  test.describe('as an admin', () => {
    test.use({ role: 'admin' })

    test('the User Management page loads both Lecturer and Editor tables', async ({
      page,
    }) => {
      const userMgmt = new CourseUserManagementPage(
        page,
        SEEDED_COURSES.iPraktikum.id,
      )
      await userMgmt.goto()
      await userMgmt.expectLoaded()
    })
  })

  test.describe('as a student', () => {
    test.use({ role: 'student' })

    test('the User Management page is blocked', async ({ page }) => {
      const userMgmt = new CourseUserManagementPage(
        page,
        SEEDED_COURSES.iPraktikum.id,
      )
      await userMgmt.goto()
      await userMgmt.expectBlocked()
    })
  })
})
