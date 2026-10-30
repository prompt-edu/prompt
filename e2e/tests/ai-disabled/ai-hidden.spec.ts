import { SEEDED_COURSES } from '../../src/data/constants'
import { BASE_URL } from '../../src/env'
import { expect, test } from '../../src/fixtures/auth'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'

// Runs in the core shard, whose stack keeps AI_ENABLED off and has no AI server.
test.describe('AI switched off', () => {
  test.use({ role: 'admin' })

  test('no AI server answers', async ({ request }) => {
    const res = await request.get(`${BASE_URL}/ai/api/info`)
    expect(res.ok()).toBe(false)
  })

  test('no AI surface is shown', async ({ page }) => {
    const remoteEntryRequests: string[] = []
    page.on('request', (request) => {
      if (request.url().includes('/ai/remoteEntry')) remoteEntryRequests.push(request.url())
    })
    const settings = new CourseSettingsPage(page, SEEDED_COURSES.fullCourse.id)
    await settings.goto()
    await settings.expectLoaded()
    await expect(page.getByRole('button', { name: 'Danger Zone' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Audit Log', exact: true })).toBeVisible()

    await expect(page.getByRole('button', { name: 'AI Calls', exact: true })).toHaveCount(0)
    expect(remoteEntryRequests, 'the AI remote is not even registered while AI is off').toEqual([])
  })
})
