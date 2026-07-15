import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'
import { cleanupCourses, createCourseViaApi, getCourseById, uniqueSuffix } from './helpers'

test.use({ role: 'admin' })

test.describe('course edit metadata', () => {
  let courseId: string
  let runToken: string

  test.beforeAll(async ({}, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    runToken = suffix
    const admin = await apiContextFor('admin')
    try {
      courseId = await createCourseViaApi(admin, {
        name: `E2EEdit${suffix}`,
        semesterTag: `e2eedit${suffix}`,
      })
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('admin edits the short description and it persists', async ({ page }) => {
    const updated = `Updated by e2e ${runToken}`

    const settings = new CourseSettingsPage(page, courseId)
    await settings.goto()
    await settings.expectLoaded()
    await settings.editShortDescription(updated)

    const api = await apiContextFor('admin')
    try {
      const course = await getCourseById(api, courseId)
      expect(course.shortDescription).toBe(updated)
    } finally {
      await api.dispose()
    }
  })
})
