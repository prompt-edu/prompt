import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CourseUserManagementPage } from '../../src/pages/CourseUserManagementPage'
import { SEEDED_COURSES } from '../../src/data/constants'
import { cleanupCourses, createCourseViaApi, uniqueSuffix } from '../courses/helpers'

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

  test.describe('as a lecturer of the course', () => {
    test.use({ role: 'course-lecturer' })

    test('the User Management page loads', async ({ page }) => {
      const userMgmt = new CourseUserManagementPage(
        page,
        SEEDED_COURSES.iPraktikum.id,
      )
      await userMgmt.goto()
      await userMgmt.expectLoaded()
    })
  })

  test.describe('as an editor of the course', () => {
    test.use({ role: 'course-editor' })

    test('the User Management page is blocked', async ({ page }) => {
      const userMgmt = new CourseUserManagementPage(
        page,
        SEEDED_COURSES.fullCourse.id,
      )
      await userMgmt.goto()
      await userMgmt.expectBlocked()
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

// The seeded courses have no staff groups in Keycloak, so these run on a course
// created through the API: creating a course provisions its Lecturer and Editor
// groups and makes the creator a Lecturer.
test.describe('course user management on a provisioned course', () => {
  test.use({ role: 'admin' })

  let courseId: string

  test.beforeAll(async () => {
    const admin = await apiContextFor('admin')
    try {
      const suffix = uniqueSuffix(test.info().workerIndex)
      courseId = await createCourseViaApi(admin, {
        name: `E2EStaffUi${suffix}`,
        semesterTag: `e2estaffui${suffix}`,
      })
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('the course creator is listed as a Lecturer and cannot remove themselves', async ({
    page,
  }) => {
    const userMgmt = new CourseUserManagementPage(page, courseId)
    await userMgmt.goto()
    await userMgmt.expectLoaded()

    await expect(userMgmt.removeButton('Lecturer', 'admin')).toBeDisabled()
  })

  test('an admin adds an Editor and removes them again', async ({ page }) => {
    const userMgmt = new CourseUserManagementPage(page, courseId)
    await userMgmt.goto()
    await userMgmt.expectLoaded()
    await expect(page.getByText('No editors yet.')).toBeVisible()

    await userMgmt.addMember('Editor', 'course-editor')
    await expect(userMgmt.removeButton('Editor', 'course-editor')).toBeVisible()

    await userMgmt.removeMember('Editor', 'course-editor')
    await expect(page.getByText('No editors yet.')).toBeVisible()
  })
})
