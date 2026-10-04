import { apiContextFor } from '../../src/fixtures/api'
import { test, expect } from '../../src/fixtures/auth'
import { SEEDED_COURSES } from '../../src/data/constants'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'

// Runs in the core shard, whose stack keeps AI_ENABLED off and has no AI server.
test.describe('AI switched off', () => {
  test.use({ role: 'admin' })

  test('core reports AI as disabled', async () => {
    const admin = await apiContextFor('admin')
    try {
      const res = await admin.get('/api/ai/status')
      expect(res.status()).toBe(200)
      expect(await res.json()).toEqual({ enabled: false })
    } finally {
      await admin.dispose()
    }
  })

  test('no AI surface is shown', async ({ page }) => {
    const settings = new CourseSettingsPage(page, SEEDED_COURSES.fullCourse.id)
    await settings.goto()
    await settings.expectLoaded()
    await expect(page.getByRole('button', { name: 'Danger Zone' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Audit Log', exact: true })).toBeVisible()

    await expect(settings.aiCard()).toHaveCount(0)
    await expect(page.getByRole('button', { name: 'AI Calls', exact: true })).toHaveCount(0)
  })
})
