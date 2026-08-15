import { test, expect } from '../../src/fixtures/auth'
import { apiContextFor } from '../../src/fixtures/api'
import { CopyCoursePage } from '../../src/pages/CopyCoursePage'
import {
  cleanupCourses,
  createCourseViaApi,
  listCourses,
  listTemplates,
  uniqueSuffix,
} from './helpers'

test.use({ role: 'admin' })

test.describe('course copy from template', () => {
  let templateId: string | undefined
  let copyId: string | undefined
  let templateName: string

  test.beforeAll(async ({}, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    templateName = `E2ECopySource${suffix}`
    const admin = await apiContextFor('admin')
    try {
      templateId = await createCourseViaApi(admin, {
        name: templateName,
        semesterTag: `e2ecopysrc${suffix}`,
        template: true,
      })
      // The template must appear in the "Use Template" picker.
      const templates = await listTemplates(admin)
      expect(templates.some((c) => c.id === templateId)).toBeTruthy()
    } finally {
      await admin.dispose()
    }
  })

  test.afterAll(async () => {
    await cleanupCourses(templateId, copyId)
  })

  test('admin copies a course from a template', async ({ page }, testInfo) => {
    const suffix = uniqueSuffix(testInfo.workerIndex)
    const targetName = `E2ECopyTarget${suffix}`

    const copy = new CopyCoursePage(page)
    copyId = await copy.copyFromTemplate({
      templateName,
      name: targetName,
      semesterTag: `e2ecopytgt${suffix}`,
    })

    const api = await apiContextFor('admin')
    try {
      const all = await listCourses(api)
      expect(all.some((c) => c.id === copyId && c.name === targetName)).toBeTruthy()
    } finally {
      await api.dispose()
    }
  })
})
