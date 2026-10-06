import { apiContextFor } from '../../src/fixtures/api'
import { test, expect } from '../../src/fixtures/auth'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { AICallsPage } from '../../src/pages/AICallsPage'
import { complete, SUMMARY_ANSWER } from './helpers'

const PHASE = FULL_COURSE_PHASES.assessment

test.describe('AI calls page', () => {
  test.beforeAll(async () => {
    const lecturer = await apiContextFor('course-lecturer')
    try {
      expect((await complete(lecturer, PHASE.id)).status()).toBe(200)
    } finally {
      await lecturer.dispose()
    }
  })

  test.describe('as an admin', () => {
    test.use({ role: 'admin' })

    test('lists the call and opens its content', async ({ page }) => {
      const aiCalls = new AICallsPage(page)
      await aiCalls.goto(SEEDED_COURSES.fullCourse.id)
      await aiCalls.expectLoaded()
      await expect(page.getByRole('button', { name: 'AI Calls', exact: true })).toBeVisible()
      await aiCalls.selectPhase(PHASE.type)

      const row = aiCalls.rows.filter({ hasText: 'success' }).first()
      await expect(row).toContainText('success')
      await row.getByRole('button', { name: 'Open' }).click()

      const dialog = page.getByRole('dialog')
      await expect(dialog.getByText('AI-generated')).toBeVisible()
      await expect(dialog.getByText(SUMMARY_ANSWER)).toBeVisible()
      await expect(dialog.getByText(/content_viewed/)).toBeVisible()
    })
  })

  test.describe('as a course lecturer', () => {
    test.use({ role: 'course-lecturer' })

    test('is not offered the page', async ({ page }) => {
      const aiCalls = new AICallsPage(page)
      await aiCalls.goto(SEEDED_COURSES.fullCourse.id)
      await expect(aiCalls.accessDenied).toBeVisible({ timeout: 15_000 })
      await expect(page.getByRole('button', { name: 'AI Calls', exact: true })).toHaveCount(0)
    })
  })
})
