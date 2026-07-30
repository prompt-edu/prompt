import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { CORE_API_URL } from '../../src/env'
import { FULL_COURSE_PHASES } from '../../src/data/constants'

const PHASE_ID = FULL_COURSE_PHASES.matching.id
const participationsPath = `/api/course_phases/${PHASE_ID}/participations`

test.describe('matching API auth (core participations)', () => {
  test('rejects unauthenticated requests', async () => {
    const anon = await request.newContext()
    try {
      const res = await anon.get(`${CORE_API_URL}${participationsPath}`)
      expect(res.status()).toBe(401)
    } finally {
      await anon.dispose()
    }
  })

  test('rejects a student on the participation list', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.get(participationsPath)
    expect(res.status()).toBe(403)
  })

  test('rejects a student on the batch update', async ({ apiAs }) => {
    const api = await apiAs('student')
    const res = await api.put(participationsPath, { data: [] })
    expect(res.status()).toBe(403)
  })

  test('rejects a course editor on the batch update', async ({ apiAs }) => {
    const api = await apiAs('course-editor')
    const res = await api.put(participationsPath, { data: [] })
    expect(res.status()).toBe(403)
  })

  test('accepts a course editor on the participation list', async ({ apiAs }) => {
    const api = await apiAs('course-editor')
    const res = await api.get(participationsPath)
    expect(res.status()).toBe(200)
  })

  test('accepts a lecturer on the participation list', async ({ apiAs }) => {
    const api = await apiAs('course-lecturer')
    const res = await api.get(participationsPath)
    expect(res.status()).toBe(200)
  })
})
