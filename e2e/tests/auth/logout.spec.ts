import { test } from '@playwright/test'
import { LoginPage } from '../../src/pages/LoginPage'
import { AuthenticatedApplicationPage } from '../../src/pages/AuthenticatedApplicationPage'
import { ROLES } from '../../src/data/roles'
import { FULL_COURSE_PHASES } from '../../src/data/constants'

// Logs in fresh instead of reusing a global-setup storageState: logging out ends
// the Keycloak session, which would invalidate a session other specs share.
test.use({ storageState: { cookies: [], origins: [] } })

test.describe('Keycloak logout flow', () => {
  test('logging out of the management console ends the Keycloak session', async ({
    page,
  }) => {
    const login = new LoginPage(page)
    await login.gotoAndLogin(ROLES.admin)

    await login.logoutFromUserMenu()
    await login.expectLoggedOut()

    // Logging in again must ask for credentials instead of reusing the SSO session.
    await login.clickLoginButton()
    await login.expectLoginFormVisible()
  })

  test('logging out of the application page ends the Keycloak session', async ({
    page,
  }) => {
    const login = new LoginPage(page)
    await login.gotoAndLogin(ROLES.student)

    const application = new AuthenticatedApplicationPage(page)
    await application.goto(FULL_COURSE_PHASES.application.id)
    await application.expectLoaded()

    await login.logoutFromUserMenu({ confirm: true })
    await login.expectLoggedOut()

    await login.goto()
    await login.expectLoginFormVisible()
  })
})
