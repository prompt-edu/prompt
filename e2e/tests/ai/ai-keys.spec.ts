import { test, expect } from '../../src/fixtures/auth'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'

const PHASE = 'Team Allocation'

test.describe('AI keys in the course settings', () => {
  test.use({ role: 'course-lecturer' })

  test('a lecturer sets, rotates and removes a phase key', async ({ page }) => {
    expect(FULL_COURSE_PHASES.teamAllocation.type).toBe(PHASE)
    const settings = new CourseSettingsPage(page, SEEDED_COURSES.fullCourse.id)
    await settings.goto()
    await settings.expectLoaded()
    await settings.expandAI()
    const row = settings.keyRow(PHASE)
    await expect(row).toContainText('No key, so AI is off for this phase.')

    await settings.saveKey(PHASE, 'logos-e2e-key-first-1111')
    await expect(row).toContainText('Key ending in 1111')
    await expect(page.getByText('logos-e2e-key-first-1111')).toHaveCount(0)

    await settings.saveKey(PHASE, 'logos-e2e-key-rotated-2222')
    await expect(row).toContainText('Key ending in 2222')

    await settings.removeKey(PHASE)
    await expect(row).toContainText('No key, so AI is off for this phase.')
  })
})
