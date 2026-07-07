import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { CORE_API_URL } from '../../src/env'
import { SEEDED_COURSE_COUNT, SEEDED_COURSES } from '../../src/data/constants'

interface Course {
  id: string
  name: string
  semesterTag: string
}

test.describe('core API: courses', () => {
  test('GET /api/courses returns the seeded courses for an admin', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/courses/')
    expect(res.ok()).toBeTruthy()

    const courses = (await res.json()) as Course[]
    expect(Array.isArray(courses)).toBeTruthy()
    expect(courses.length).toBeGreaterThanOrEqual(SEEDED_COURSE_COUNT)

    const names = courses.map((c) => c.name)
    expect(names).toContain(SEEDED_COURSES.iPraktikum.name)
  })

  test('GET /api/courses without a token is rejected', async () => {
    const anon = await request.newContext({ baseURL: CORE_API_URL })
    const res = await anon.get('/api/courses/')
    expect(res.status()).toBe(401)
    await anon.dispose()
  })

  test('GET /api/course_phase_types is publicly readable and non-empty', async ({
    apiAs,
  }) => {
    // This endpoint requires no auth, but any valid context works.
    const api = await apiAs('student')
    const res = await api.get('/api/course_phase_types')
    expect(res.ok()).toBeTruthy()

    const types = (await res.json()) as unknown[]
    expect(Array.isArray(types)).toBeTruthy()
    expect(types.length).toBeGreaterThan(0)
  })
})
