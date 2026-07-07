import fs from 'fs'
import path from 'path'
import { chromium, FullConfig, request } from '@playwright/test'
import { ROLES, SEEDED_ROLES } from './data/roles'
import { LoginPage } from './pages/LoginPage'
import { authFile } from './fixtures/auth'
import { BASE_URL, CORE_API_URL, SELF_TEAM_ALLOCATION_API } from './env'

// Poll a URL until it responds (any HTTP status), or time out. The core server
// image is distroless (no container healthcheck), so we gate readiness here.
async function waitForUrl(url: string, label: string, timeoutMs = 120_000) {
  const ctx = await request.newContext()
  const deadline = Date.now() + timeoutMs
  try {
    for (;;) {
      try {
        const res = await ctx.get(url, { timeout: 5_000 })
        if (res.status() > 0) {
          // eslint-disable-next-line no-console
          console.log(`[global-setup] ${label} ready (${res.status()})`)
          return
        }
      } catch {
        // not up yet
      }
      if (Date.now() > deadline) {
        throw new Error(`[global-setup] timed out waiting for ${label} at ${url}`)
      }
      await new Promise((r) => setTimeout(r, 2_000))
    }
  } finally {
    await ctx.dispose()
  }
}

// Poll a phase server's /info endpoint THROUGH the client-core proxy until it
// returns the expected service JSON. A plain status check is not enough: the
// SPA fallback answers 200 with index.html for misrouted paths, so this also
// proves the nginx proxy wiring before any spec runs.
async function waitForServiceInfo(url: string, serviceName: string, timeoutMs = 120_000) {
  const ctx = await request.newContext()
  const deadline = Date.now() + timeoutMs
  try {
    for (;;) {
      try {
        const res = await ctx.get(url, { timeout: 5_000 })
        if (res.ok()) {
          const info = (await res.json()) as { serviceName?: string; healthy?: boolean }
          if (info.serviceName === serviceName && info.healthy === true) {
            // eslint-disable-next-line no-console
            console.log(`[global-setup] ${serviceName} ready (via proxy)`)
            return
          }
        }
      } catch {
        // not up yet, or the SPA fallback answered with HTML
      }
      if (Date.now() > deadline) {
        throw new Error(`[global-setup] timed out waiting for ${serviceName} at ${url}`)
      }
      await new Promise((r) => setTimeout(r, 2_000))
    }
  } finally {
    await ctx.dispose()
  }
}

// Logs in once per seeded role through the REAL Keycloak login form and saves
// the resulting browser session (localStorage JWT + Keycloak SSO cookies) to a
// storageState file. Tests reuse these, so the auth flow is exercised once here
// and runs fast afterwards. A Keycloak SSO cookie (24h) keeps sessions valid
// even after the 5-min access token expires; the app silently re-authenticates.
export default async function globalSetup(_config: FullConfig) {
  const authDir = path.resolve(__dirname, '../.auth')
  fs.mkdirSync(authDir, { recursive: true })

  // Wait for the stack to be reachable before driving the login flow.
  await waitForUrl(`${CORE_API_URL}/api/hello`, 'core server')
  await waitForUrl(BASE_URL, 'core client')
  await waitForServiceInfo(
    `${BASE_URL}${SELF_TEAM_ALLOCATION_API}/info`,
    'self-team-allocation',
  )

  const browser = await chromium.launch()
  try {
    for (const role of SEEDED_ROLES) {
      const context = await browser.newContext()
      const page = await context.newPage()
      const login = new LoginPage(page)

      try {
        await login.gotoAndLogin(ROLES[role])
        await context.storageState({ path: authFile(role) })
        // eslint-disable-next-line no-console
        console.log(`[global-setup] authenticated role: ${role}`)
      } catch (err) {
        // Capture what the browser actually landed on, to debug auth/redirects.
        try {
          const shot = path.resolve(__dirname, `../test-results/global-setup-${role}.png`)
          await page.screenshot({ path: shot, fullPage: true })
          const title = await page.title().catch(() => '<no title>')
          const body = (await page.locator('body').innerText().catch(() => '')).slice(0, 800)
          // eslint-disable-next-line no-console
          console.error(
            `[global-setup] DIAG role=${role} url=${page.url()} title=${JSON.stringify(title)}\n` +
              `[global-setup] body:\n${body}`,
          )
        } catch {
          /* best-effort diagnostics */
        }
        throw new Error(
          `[global-setup] failed to authenticate role "${role}": ${String(err)}`,
        )
      } finally {
        await context.close()
      }
    }
  } finally {
    await browser.close()
  }
}
