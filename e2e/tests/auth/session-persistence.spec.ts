import { test, expect } from '../../src/fixtures/auth'
import { FULL_COURSE_PHASES } from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.application.id

// A student with a stored session must be able to cross between the public
// application site and the app without ever going back through Keycloak: the
// session is restored from the refresh token, not from a redirect.
test.use({ role: 'student' })

test.describe('auth: session persistence between apply and management', () => {
  test('a logged-in student switches between apply and management without re-authenticating', async ({
    page,
  }) => {
    const authRedirects: string[] = []
    let refreshGrants = 0
    page.on('request', (request) => {
      const url = request.url()
      if (url.includes('/protocol/openid-connect/auth')) {
        authRedirects.push(url)
      }
      if (
        url.includes('/protocol/openid-connect/token') &&
        request.postData()?.includes('grant_type=refresh_token')
      ) {
        refreshGrants += 1
      }
    })

    // Public apply entry point: the session is restored and the login card skipped.
    await page.goto(`/apply/${PHASE_ID}`)
    await expect(page).toHaveURL(new RegExp(`/apply/${PHASE_ID}/authenticated$`))
    await expect(page.getByRole('heading', { name: 'Personal Information' })).toBeVisible()
    expect(refreshGrants).toBeGreaterThan(0)

    // Into the app, and back to the application site: authenticated both ways.
    await page.getByRole('button', { name: 'Go to App' }).click()
    await expect(page).toHaveURL(/\/management/)

    await page.goto(`/apply/${PHASE_ID}`)
    await expect(page).toHaveURL(new RegExp(`/apply/${PHASE_ID}/authenticated$`))
    await expect(page.getByRole('heading', { name: 'Personal Information' })).toBeVisible()

    // A full reload of a protected route restores the session too.
    await page.goto('/management')
    await expect(page).toHaveURL(/\/management/)
    await page.reload()
    await expect(page).toHaveURL(/\/management/)

    await expect(page.locator('#kc-login')).toHaveCount(0)
    expect(authRedirects).toEqual([])
  })
})
