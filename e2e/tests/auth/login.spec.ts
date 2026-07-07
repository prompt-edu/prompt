import { test, expect } from '@playwright/test'
import { LoginPage } from '../../src/pages/LoginPage'
import { ROLES } from '../../src/data/roles'

// These tests exercise the REAL login flow, so they must start without a stored
// session. (Browser specs elsewhere reuse the storageState from global-setup.)
test.use({ storageState: { cookies: [], origins: [] } })

test.describe('Keycloak login flow', () => {
  test('unauthenticated visit redirects to the Keycloak login form', async ({
    page,
  }) => {
    const login = new LoginPage(page)
    await login.goto()
    await login.expectLoginFormVisible()
  })

  test('admin can log in and lands in the management console', async ({
    page,
  }) => {
    const login = new LoginPage(page)
    await login.gotoAndLogin(ROLES.admin)
    await expect(page).toHaveURL(/\/management/)
  })

  test('wrong password is rejected', async ({ page }) => {
    const login = new LoginPage(page)
    await login.goto()
    await login.login({ ...ROLES.student, password: 'definitely-wrong' })
    // Keycloak re-renders the form with an error; we never reach /management.
    await expect(page.locator('#kc-login')).toBeVisible()
    await expect(page).not.toHaveURL(/\/management/)
  })
})
