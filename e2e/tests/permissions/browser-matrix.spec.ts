import type { Page } from '@playwright/test'
import { test, expect } from '../../src/fixtures/auth'
import {
  MATRIX_ROLES,
  PRIMARY_COURSE,
  SURFACES,
  CROSS_COURSE,
} from '../../src/data/permissionMatrix'

const browserSurfaces = SURFACES.filter((s) => s.browser)

async function expectAllowed(page: Page, heading: string, assertText?: string) {
  await expect(page.getByRole('heading', { level: 1, name: heading })).toBeVisible()
  if (assertText) {
    await expect(page.getByText(assertText).first()).toBeVisible()
  }
}

// Blocked pages render UnauthorizedPage ("Access Denied"). Asserting that overlay
// (not just the missing heading) rules out a false pass from a slow load, a wrong
// route, or a crash.
async function expectBlocked(page: Page, heading: string) {
  await expect(page.getByText('Access Denied')).toBeVisible()
  await expect(page.getByRole('heading', { level: 1, name: heading })).toBeHidden()
}

test.describe('permission matrix (browser)', () => {
  for (const role of MATRIX_ROLES) {
    test.describe(role, () => {
      test.use({ role })

      for (const surface of browserSurfaces) {
        const browser = surface.browser!
        const allowed = browser.allowed.includes(role)

        test(`${allowed ? 'sees' : 'is blocked from'} ${surface.name}`, async ({ page }) => {
          await page.goto(browser.path(PRIMARY_COURSE.id))
          if (allowed) await expectAllowed(page, browser.heading, browser.assertText)
          else await expectBlocked(page, browser.heading)
        })
      }
    })
  }
})

test.describe('cross-course isolation (browser)', () => {
  for (const testCase of CROSS_COURSE) {
    test.describe(`${testCase.role} cannot reach ${testCase.course.name}`, () => {
      test.use({ role: testCase.role })

      for (const surfaceName of testCase.surfaces) {
        const surface = SURFACES.find((s) => s.name === surfaceName)
        if (!surface?.browser) continue
        const browser = surface.browser

        test(`is blocked from ${surfaceName}`, async ({ page }) => {
          await page.goto(browser.path(testCase.course.id))
          await expectBlocked(page, browser.heading)
        })
      }
    })
  }
})
