import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'
import { cleanupCourses, createCourseViaApi, listCourses, uniqueSuffix } from './helpers'

test.use({ role: 'admin' })

test.describe('course delete', () => {
  let courseId: string | undefined

  test.beforeAll(async ({}, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    const admin = await apiContextFor('admin')
    try {
      courseId = await createCourseViaApi(admin, {
        name: `E2EDelete${suffix}`,
        semesterTag: `e2edelete${suffix}`,
      })
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    // Best-effort in case the UI delete failed mid-test.
    await cleanupCourses(courseId)
  })

  test('admin deletes a course', async ({ page }) => {
    const id = courseId!

    const settings = new CourseSettingsPage(page, id)
    await settings.goto()
    await settings.expectLoaded()
    await settings.deleteCourse()

    const api = await apiContextFor('admin')
    try {
      const all = await listCourses(api)
      expect(all.some((c) => c.id === id)).toBeFalsy()
    } finally {
      await api.dispose()
    }
    courseId = undefined // deleted; skip the afterAll double-delete
  })
})
