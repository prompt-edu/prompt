import type { APIRequestContext } from '@playwright/test'
import { apiContextFor } from '../../src/fixtures/api'
import { expect, test } from '../../src/fixtures/auth'
import {
  SEEDED_WELCOME_TEXT,
  WELCOME_TEXT_COURSE_ID,
  WELCOME_TEXT_PHASE_ID,
} from '../../src/data/constants'
import { ApplicationAdminPage } from '../../src/pages/ApplicationAdminPage'
import { ApplyPage } from '../../src/pages/ApplyPage'

const EDITED_TEXT = 'Applying to the right course? This is iPraktikumWelcome.'

// Restoring through the API rather than the editor, because typing into the rich
// text editor could never reproduce the seeded markup.
async function setWelcomeTextViaApi(api: APIRequestContext, welcomeText: string | null) {
  const res = await api.put(`/api/course_phases/${WELCOME_TEXT_PHASE_ID}`, {
    data: { id: WELCOME_TEXT_PHASE_ID, restrictedData: { welcomeText } },
  })
  expect(res.ok()).toBeTruthy()
}

test.describe('application: instructor welcome text', () => {
  test.describe('as an applicant', () => {
    // The apply page is public; a signed-in session would be redirected to the
    // authenticated variant instead.
    test.use({ storageState: { cookies: [], origins: [] } })

    test('the welcome text renders above the form, sanitized', async ({ page }) => {
      const apply = new ApplyPage(page)
      await apply.goto(WELCOME_TEXT_PHASE_ID)

      await expect(apply.welcomeText).toContainText(SEEDED_WELCOME_TEXT.paragraph, {
        timeout: 15_000,
      })

      // The DOMPurify hook forces external links to open safely.
      const link = apply.welcomeLink(SEEDED_WELCOME_TEXT.linkName)
      await expect(link).toHaveAttribute('href', SEEDED_WELCOME_TEXT.linkHref)
      await expect(link).toHaveAttribute('target', '_blank')
      await expect(link).toHaveAttribute('rel', 'noopener')

      // The seeded HTML also carries a script tag and an inline event handler.
      // Neither may survive, and neither may ever have run.
      await expect(apply.welcomeText.locator('script')).toHaveCount(0)
      await expect(apply.welcomeText.locator('[onerror]')).toHaveCount(0)
      expect(await apply.readGlobal(SEEDED_WELCOME_TEXT.xssMarker)).toBeUndefined()
    })
  })

  test.describe('as an instructor', () => {
    test.use({ role: 'admin' })

    let api: APIRequestContext

    test.beforeAll(async () => {
      api = await apiContextFor('admin')
    })

    // Put the seeded markup back so the applicant case above does not depend on
    // whether this one ran first.
    test.afterAll(async () => {
      await setWelcomeTextViaApi(api, SEEDED_WELCOME_TEXT.html)
      await api.dispose()
    })

    test('an instructor edits, publishes and clears the welcome text', async ({ page, browser }) => {
      const admin = new ApplicationAdminPage(page)
      await admin.gotoSettings(WELCOME_TEXT_COURSE_ID, WELCOME_TEXT_PHASE_ID)

      await expect(admin.welcomeTextEditor).toContainText(SEEDED_WELCOME_TEXT.paragraph, {
        timeout: 15_000,
      })

      await admin.setWelcomeText(EDITED_TEXT)

      // It survives a reload, so it was persisted and not merely held in state.
      await admin.gotoSettings(WELCOME_TEXT_COURSE_ID, WELCOME_TEXT_PHASE_ID)
      await expect(admin.welcomeTextEditor).toContainText(EDITED_TEXT, { timeout: 15_000 })

      const published = await browser.newContext({ storageState: { cookies: [], origins: [] } })
      try {
        const apply = new ApplyPage(await published.newPage())
        await apply.goto(WELCOME_TEXT_PHASE_ID)
        await expect(apply.welcomeText).toContainText(EDITED_TEXT, { timeout: 15_000 })
      } finally {
        await published.close()
      }

      // Clearing the editor removes the text from the applicant's page entirely.
      await admin.setWelcomeText('')

      const cleared = await browser.newContext({ storageState: { cookies: [], origins: [] } })
      try {
        const apply = new ApplyPage(await cleared.newPage())
        await apply.goto(WELCOME_TEXT_PHASE_ID)
        await apply.expectCourseHeading('iPraktikumWelcome')
        await expect(apply.welcomeText).toHaveCount(0)
      } finally {
        await cleared.close()
      }
    })
  })
})
