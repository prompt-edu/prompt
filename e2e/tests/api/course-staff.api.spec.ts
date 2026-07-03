import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { tokenFor } from '../../src/fixtures/api'
import { CORE_API_URL } from '../../src/env'
import { SEEDED_COURSES } from '../../src/data/constants'

// Course-staff (Lecturer / Editor) management endpoints. The seeded e2e realm
// does not create per-course /Prompt/<courseGroupName>/{Lecturer,Editor}
// subgroups, so happy-path calls hit the "subgroup missing -> empty list"
// branch. That is intentional: the endpoint must degrade to empty tables
// instead of 500-ing when a course has no staff groups yet, and this suite
// pins that behavior.

interface StaffMember {
  keycloakUserID: string
  username: string
  email: string
  firstName: string
  lastName: string
}

interface CourseStaff {
  lecturers: StaffMember[]
  editors: StaffMember[]
}

interface UserSearchResults {
  results: StaffMember[]
  truncated: boolean
}

const COURSE_ID = SEEDED_COURSES.iPraktikum.id

// Decode the `sub` (Keycloak user ID) from a JWT without a verifying library.
function subjectFromJwt(jwt: string): string {
  const payload = jwt.split('.')[1]
  const json = Buffer.from(payload, 'base64').toString('utf8')
  return (JSON.parse(json) as { sub: string }).sub
}

test.describe('core API: keycloak course-staff', () => {
  test('GET /keycloak/:courseID/group/staff returns both groups as an admin', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.get(`/api/keycloak/${COURSE_ID}/group/staff`)
    expect(res.ok()).toBeTruthy()

    const staff = (await res.json()) as CourseStaff
    expect(Array.isArray(staff.lecturers)).toBeTruthy()
    expect(Array.isArray(staff.editors)).toBeTruthy()
  })

  test('GET /keycloak/:courseID/group/staff is forbidden for a student', async ({
    apiAs,
  }) => {
    const api = await apiAs('student')
    const res = await api.get(`/api/keycloak/${COURSE_ID}/group/staff`)
    expect(res.status()).toBe(403)
  })

  test('GET /keycloak/:courseID/group/staff without a token is rejected', async () => {
    const anon = await request.newContext({ baseURL: CORE_API_URL })
    const res = await anon.get(`/api/keycloak/${COURSE_ID}/group/staff`)
    expect(res.status()).toBe(401)
    await anon.dispose()
  })

  test('PUT /keycloak/:courseID/group/:groupName/members rejects a group name outside the allow-list', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.put(
      `/api/keycloak/${COURSE_ID}/group/Admin/members/00000000-0000-0000-0000-000000000000`,
    )
    expect(res.status()).toBe(400)
    const body = (await res.json()) as { error: string }
    expect(body.error).toMatch(/invalid group name/i)
  })

  test('DELETE /keycloak/:courseID/group/:groupName/members rejects self-removal', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const adminSub = subjectFromJwt(await tokenFor('admin'))
    const res = await api.delete(
      `/api/keycloak/${COURSE_ID}/group/Lecturer/members/${adminSub}`,
    )
    expect(res.status()).toBe(400)
    const body = (await res.json()) as { error: string }
    expect(body.error).toMatch(/cannot remove yourself/i)
  })
})

test.describe('core API: keycloak user search', () => {
  test('GET /keycloak/users/search returns matching users for an admin', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/keycloak/users/search?q=admin')
    expect(res.ok()).toBeTruthy()

    const results = (await res.json()) as UserSearchResults
    expect(Array.isArray(results.results)).toBeTruthy()
    expect(results.results.length).toBeGreaterThan(0)
    expect(results.results.some((u) => u.username === 'admin')).toBeTruthy()
  })

  test('GET /keycloak/users/search rejects a query shorter than 2 characters', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/keycloak/users/search?q=a')
    expect(res.status()).toBe(400)
    const body = (await res.json()) as { error: string }
    expect(body.error).toMatch(/at least 2 characters/i)
  })

  test('GET /keycloak/users/search rejects a non-integer limit', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/keycloak/users/search?q=admin&limit=abc')
    expect(res.status()).toBe(400)
    const body = (await res.json()) as { error: string }
    expect(body.error).toMatch(/limit must be an integer/i)
  })

  test('GET /keycloak/users/search is forbidden for a student', async ({
    apiAs,
  }) => {
    const api = await apiAs('student')
    const res = await api.get('/api/keycloak/users/search?q=admin')
    expect(res.status()).toBe(403)
  })
})
