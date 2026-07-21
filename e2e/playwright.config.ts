import { defineConfig, devices } from '@playwright/test'
import { BASE_URL } from './src/env'

const isCI = !!process.env.CI

// Worker count. In CI each module shard runs in its own stack (see
// .github/workflows/test-e2e.yml), so a single worker per shard already isolates
// modules from each other; PW_WORKERS can override it once specs are audited for
// intra-stack parallelism. Locally, undefined lets Playwright pick.
const workers = process.env.PW_WORKERS
  ? Number(process.env.PW_WORKERS)
  : isCI
    ? 1
    : undefined

// Runtime budget guard: fail the run if a shard exceeds the agreed wall-clock
// budget (default 20 min in CI) instead of silently creeping up. 0 disables it.
const budgetMin = Number(process.env.PW_GLOBAL_TIMEOUT_MIN ?? (isCI ? 20 : 0))

export default defineConfig({
  testDir: './tests',
  // Login flows mutate the shared Keycloak session, and seeded data is shared;
  // keep within-file order deterministic but allow files to run in parallel.
  fullyParallel: false,
  forbidOnly: isCI,
  retries: isCI ? 2 : 0,
  workers,
  timeout: 60_000,
  globalTimeout: budgetMin > 0 ? budgetMin * 60_000 : undefined,
  expect: { timeout: 10_000 },

  // Logs in each seeded role once and writes storageState files.
  globalSetup: require.resolve('./src/global-setup'),

  // When sharded in CI each module shard emits a uniquely-named blob report
  // (PW_BLOB=1 + PW_SHARD_NAME=<module>); a downstream job merges them into one
  // HTML report. Otherwise (local, or an unsharded run) produce HTML directly.
  reporter:
    process.env.PW_BLOB === '1'
      ? [
          [
            'blob',
            {
              outputDir: 'blob-report',
              fileName: `report-${process.env.PW_SHARD_NAME ?? 'shard'}.zip`,
            },
          ] as const,
        ]
      : [
          ['list'] as const,
          ['html', { outputFolder: 'playwright-report', open: 'never' }] as const,
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
