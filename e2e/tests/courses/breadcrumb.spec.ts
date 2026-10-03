import { test, expect } from '../../src/fixtures/auth'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { FULL_COURSE_PHASES, SEEDED_COURSES } from '../../src/data/constants'

const WIDE = { width: 1280, height: 800 }
// Leaves the header too little room for "iPraktikumFull › Application › Participants".
const NARROW = { width: 320, height: 800 }

test.use({ role: 'lecturer' })

test.describe('header breadcrumb', () => {
  test('a trail that does not fit collapses into a menu and expands again', async ({ page }) => {
    await page.setViewportSize(WIDE)
    const admin = new ApplicationAdminPage(page)
    await admin.gotoParticipants(SEEDED_COURSES.fullCourse.id, FULL_COURSE_PHASES.application.id)

    const breadcrumb = page.getByRole('navigation', { name: 'breadcrumb' })
    const showHidden = breadcrumb.getByRole('button', { name: 'Show hidden breadcrumbs' })
    await expect(breadcrumb.getByText('Application', { exact: true })).toBeVisible()
    await expect(showHidden).toHaveCount(0)

    await page.setViewportSize(NARROW)
    await expect(showHidden).toBeVisible()
    await expect(breadcrumb.getByText('Application', { exact: true })).toHaveCount(0)

    await showHidden.click()
    await expect(page.getByRole('menuitem', { name: 'Application' })).toBeVisible()
    await page.keyboard.press('Escape')

    await page.setViewportSize(WIDE)
    await expect(showHidden).toHaveCount(0)
    await expect(breadcrumb.getByText('Application', { exact: true })).toBeVisible()
  })
})
