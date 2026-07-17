import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CourseCreationPage } from '../../src/pages/CourseCreationPage'
import { CoursesPage } from '../../src/pages/CoursesPage'
import { cleanupCourses, getCourseStaff, listCourses, uniqueSuffix } from './helpers'

test.use({ role: 'admin' })

test.describe('course create', () => {
  let courseId: string | undefined

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('admin creates a course and its course-scoped Keycloak roles are provisioned', async ({
    page,
  }, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    const name = `E2ECreate${suffix}`

    const creation = new CourseCreationPage(page)
    courseId = await creation.create({ name, semesterTag: `e2ecreate${suffix}` })

    // UI: the new course card shows up in the list.
    const courses = new CoursesPage(page)
    await courses.goto()
    await courses.expectLoaded()
    await courses.expectCourseVisible(name)

    const api = await apiContextFor('admin')
    try {
      // The course exists via the API...
      const all = await listCourses(api)
      expect(all.some((c) => c.id === courseId)).toBeTruthy()

      // ...and the creator (admin) landed in the Lecturer subgroup, proving
      // CreateCourseGroupsAndRoles provisioned the course-scoped roles.
      const staff = await getCourseStaff(api, courseId!)
      expect(staff.lecturers.some((m) => m.username === 'admin')).toBeTruthy()
    } finally {
      await api.dispose()
    }
  })
})
