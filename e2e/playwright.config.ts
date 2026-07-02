import { defineConfig, devices } from '@playwright/test'
import { BASE_URL } from './src/env'

const isCI = !!process.env.CI

export default defineConfig({
  testDir: './tests',
  // Login flows mutate the shared Keycloak session, and seeded data is shared;
  // keep within-file order deterministic but allow files to run in parallel.
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
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
