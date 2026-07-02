import { Page, expect } from '@playwright/test'
import { BASE_URL } from '../env'
import { RoleAccount } from '../data/roles'

// Drives the real Keycloak login form. The app uses keycloak-js with
// `onLoad: 'login-required'`, so visiting the app redirects to this form.
export class LoginPage {
  constructor(private readonly page: Page) {}

  // Navigate to a protected route. The public landing page ("/") does not
  // require auth; the management console mounts Keycloak with
  // `onLoad: 'login-required'`, so visiting it redirects to the login form.
  async goto() {
    await this.page.goto(`${BASE_URL}/management`)
  }

  // Standard Keycloak login page selectors (stable across KC versions).
  async login(account: RoleAccount) {
    await this.page.fill('#username', account.username)
    await this.page.fill('#password', account.password)
    await this.page.click('#kc-login')
  }

  // After a successful login the app lands on /management (possibly with a
  // Keycloak hash fragment, and possibly redirected to a sub-page). Wait for the
  // route, then for the JWT to be persisted so a captured storageState is valid.
  async expectLoggedIn() {
    await this.page.waitForURL(/\/management/, { timeout: 30_000 })
    await this.page.waitForFunction(() => !!localStorage.getItem('jwt_token'), {
      timeout: 30_000,
    })
  }

  async gotoAndLogin(account: RoleAccount) {
    await this.goto()
    await this.login(account)
    await this.expectLoggedIn()
  }

  async expectLoginFormVisible() {
    await expect(this.page.locator('#kc-login')).toBeVisible()
  }
}
