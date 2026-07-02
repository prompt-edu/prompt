import path from 'path'
import { test as base } from '@playwright/test'
import { Role } from '../data/roles'

// storageState files written by global-setup, one per role.
const AUTH_DIR = path.resolve(__dirname, '../../.auth')

export function authFile(role: Role): string {
  return path.join(AUTH_DIR, `${role}.json`)
}

// Extends the base test with a `role` option. Setting it (via
// `test.use({ role: 'lecturer' })`) loads that role's pre-authenticated
// session, so the test starts already logged in.
export const test = base.extend<{ role: Role }>({
  role: ['admin', { option: true }],
  storageState: async ({ role }, use) => {
    await use(authFile(role))
  },
})

export const expect = test.expect
