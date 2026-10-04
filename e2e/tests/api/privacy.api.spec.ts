import { request } from '@playwright/test'
import { CORE_API_URL } from '../../src/env'
import { test, expect } from '../../src/fixtures/api'

// The admin privacy console endpoints are PROMPT_Admin only. prompt-sdk answers
// 401 for a missing token and core answers 403 for a token without the role.
const ADMIN_ENDPOINTS = ['/api/privacy/admin/data-deletions', '/api/privacy/admin/data-exports']

test.describe('core API: privacy console authorization', () => {
  for (const endpoint of ADMIN_ENDPOINTS) {
    test(`rejects an unauthenticated caller on ${endpoint}`, async () => {
      const anonymous = await request.newContext({ baseURL: CORE_API_URL })
      try {
        const res = await anonymous.get(endpoint)
        expect(res.status()).toBe(401)
      } finally {
        await anonymous.dispose()
      }
    })

    test(`rejects a student on ${endpoint}`, async ({ apiAs }) => {
      const api = await apiAs('student')
      const res = await api.get(endpoint)
      expect(res.status()).toBe(403)
    })

    test(`accepts an admin on ${endpoint}`, async ({ apiAs }) => {
      const api = await apiAs('admin')
      const res = await api.get(endpoint)
      expect(res.status(), await res.text()).toBe(200)
    })
  }

  test('a student may still manage their own data', async ({ apiAs }) => {
    const api = await apiAs('student')
    for (const endpoint of ['/api/privacy/data-export', '/api/privacy/data-deletion']) {
      const res = await api.get(endpoint)
      expect([200, 204], `${endpoint} answered ${res.status()}`).toContain(res.status())
    }
  })
})
