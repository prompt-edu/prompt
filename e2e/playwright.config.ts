import { defineConfig, devices } from '@playwright/test'
import { BASE_URL } from './src/env'

const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  // Seeded data is shared, so keep within-file order deterministic but allow
  // files to run in parallel. (Logins no longer share a Keycloak session -
  // global-setup writes one storageState per role.)
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  // Single worker on CI: a few spec files still share a seeded phase (the
  // application phase, FULL_COURSE_PHASES.assessment, and the self-team-allocation
  // phase), so raising this needs those carved out into dedicated phases first.
  workers: isCI ? 1 : undefined,
  timeout: 60_000,
  expect: { timeout: 10_000 },

  // Logs in each seeded role once and writes storageState files.
  globalSetup: require.resolve('./src/global-setup'),

  reporter: [
    ['list'],
    ['html', { outputFolder: 'playwright-report', open: 'never' }],
    ...(isCI ? [['github'] as const] : []),
  ],

  outputDir: 'test-results',

  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    video: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'api',
      testMatch: /.*\.api\.spec\.ts/,
    },
    {
      name: 'chromium',
      testIgnore: /.*\.api\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
