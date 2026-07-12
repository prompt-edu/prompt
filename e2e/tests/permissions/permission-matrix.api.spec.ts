import { request } from '@playwright/test'
import type { APIRequestContext } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { CORE_API_URL } from '../../src/env'
import {
  MATRIX_ROLES,
  PRIMARY_COURSE,
  SURFACES,
  CROSS_COURSE,
  Surface,
} from '../../src/data/permissionMatrix'

const apiSurfaces = SURFACES.filter((s) => s.api)

function callSurface(api: APIRequestContext, surface: Surface, courseId: string) {
  const a = surface.api!
  const url = a.path(courseId)
  if (a.method === 'PUT') {
    return api.put(url, {
      headers: { 'content-type': 'application/json' },
      data: a.invalidBody ?? '',
    })
  }
  return api.get(url)
}

test.describe('permission matrix (API)', () => {
  for (const surface of apiSurfaces) {
    const a = surface.api!

    for (const role of MATRIX_ROLES) {
      const allowed = a.allowed.includes(role)
      const title = allowed
        ? `${role} may access ${surface.name}`
        : `${role} is blocked from ${surface.name} (403${
            a.blockedReason ? `, ${a.blockedReason}` : ''
          })`

      test(title, async ({ apiAs }) => {
        const api = await apiAs(role)
        const res = await callSurface(api, surface, PRIMARY_COURSE.id)
        if (allowed) {
          if (a.allowedStatus) expect(res.status()).toBe(a.allowedStatus)
          else expect(res.ok()).toBeTruthy()
        } else {
          expect(res.status()).toBe(403)
        }
      })
    }
  }
})

test.describe('unauthenticated requests are rejected (401)', () => {
  for (const surface of apiSurfaces) {
    test(`${surface.name} requires a token`, async () => {
      const anon = await request.newContext({ baseURL: CORE_API_URL })
      const res = await callSurface(anon, surface, PRIMARY_COURSE.id)
      expect(res.status()).toBe(401)
      await anon.dispose()
    })
  }
})

test.describe('cross-course isolation (API)', () => {
  for (const testCase of CROSS_COURSE) {
    for (const surfaceName of testCase.surfaces) {
      const surface = SURFACES.find((s) => s.name === surfaceName)
      if (!surface?.api) continue

      test(`${testCase.role} is blocked from ${surfaceName} on ${testCase.course.name} (403)`, async ({
        apiAs,
      }) => {
        const api = await apiAs(testCase.role)
        const res = await callSurface(api, surface, testCase.course.id)
        expect(res.status()).toBe(403)
      })
    }
  }
})
