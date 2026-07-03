import { test, expect } from '../../src/fixtures/auth'
import { CoursesPage } from '../../src/pages/CoursesPage'
import { SEEDED_COURSES } from '../../src/data/constants'

// Reuses the admin storageState from global-setup.
test.use({ role: 'admin' })

test.describe('course list', () => {
  test('admin sees the seeded courses', async ({ page }) => {
    const courses = new CoursesPage(page)
    await courses.goto()
    await courses.expectLoaded()

    await courses.expectCourseVisible(SEEDED_COURSES.iPraktikum.name)
    await courses.expectCourseVisible(SEEDED_COURSES.testCourse.name)
  })

  test('opening a course navigates to its overview', async ({ page }) => {
    const courses = new CoursesPage(page)
    await courses.goto()
    await courses.expectLoaded()

    await courses.openCourse(SEEDED_COURSES.iPraktikum.id)
    await expect(page).toHaveURL(
      new RegExp(`/management/course/${SEEDED_COURSES.iPraktikum.id}`),
    )
  })
})
