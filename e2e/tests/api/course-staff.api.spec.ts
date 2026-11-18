import { request } from '@playwright/test'
import { test, expect } from '../../src/fixtures/api'
import { apiContextFor, tokenFor } from '../../src/fixtures/api'
import { CORE_API_URL } from '../../src/env'
import { SEEDED_COURSES } from '../../src/data/constants'
import { cleanupCourses, createCourseViaApi, uniqueSuffix } from '../courses/helpers'

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
  universityLogin: string
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
// A well-formed Keycloak user ID that belongs to nobody.
const UNKNOWN_USER_ID = '00000000-0000-0000-0000-000000000000'

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

test.describe('core API: keycloak course-staff of a provisioned course', () => {
  let courseId: string | undefined

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('GET /keycloak/:courseID/group/staff returns the university login of each member', async ({
    apiAs,
  }, testInfo) => {
    const api = await apiAs('admin')
    const suffix = uniqueSuffix(testInfo.workerIndex)
    courseId = await createCourseViaApi(api, {
      name: `E2EStaffLogin${suffix}`,
      semesterTag: `e2estafflogin${suffix}`,
    })

    const editorSub = subjectFromJwt(await tokenFor('course-editor'))
    const addRes = await api.put(`/api/keycloak/${courseId}/group/Editor/members/${editorSub}`)
    expect(addRes.status()).toBe(204)

    const res = await api.get(`/api/keycloak/${courseId}/group/staff`)
    expect(res.ok()).toBeTruthy()
    const staff = (await res.json()) as CourseStaff

    const creator = staff.lecturers.find((m) => m.username === 'admin')
    expect(creator?.universityLogin).toBe('ad12min')

    const editor = staff.editors.find((m) => m.keycloakUserID === editorSub)
    expect(editor?.username).toBe('course-editor')
    expect(editor?.universityLogin).toBe('co69edt')
  })
})

test.describe('core API: keycloak course-staff membership changes', () => {
  let courseId: string
  let editorSub: string

  test.beforeAll(async () => {
    const admin = await apiContextFor('admin')
    try {
      const suffix = uniqueSuffix(test.info().workerIndex)
      courseId = await createCourseViaApi(admin, {
        name: `E2EStaffChange${suffix}`,
        semesterTag: `e2estaffchange${suffix}`,
      })
    } finally {
      await admin.dispose()
    }
    editorSub = subjectFromJwt(await tokenFor('course-editor'))
  })

  test.afterAll(async () => {
    await cleanupCourses(courseId)
  })

  test('DELETE /keycloak/:courseID/group/:groupName/members removes a member', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const memberUrl = `/api/keycloak/${courseId}/group/Editor/members/${editorSub}`
    expect((await api.put(memberUrl)).status()).toBe(204)

    const res = await api.delete(memberUrl)
    expect(res.status()).toBe(204)

    const staff = (await (
      await api.get(`/api/keycloak/${courseId}/group/staff`)
    ).json()) as CourseStaff
    expect(staff.editors.map((m) => m.keycloakUserID)).not.toContain(editorSub)
  })

  test('PUT /keycloak/:courseID/group/:groupName/members is idempotent', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const memberUrl = `/api/keycloak/${courseId}/group/Lecturer/members/${editorSub}`
    expect((await api.put(memberUrl)).status()).toBe(204)
    expect((await api.put(memberUrl)).status()).toBe(204)

    const staff = (await (
      await api.get(`/api/keycloak/${courseId}/group/staff`)
    ).json()) as CourseStaff
    expect(staff.lecturers.filter((m) => m.keycloakUserID === editorSub)).toHaveLength(1)
  })

  test('PUT /keycloak/:courseID/group/:groupName/members returns 404 for an unknown user', async ({
    apiAs,
  }) => {
    const api = await apiAs('admin')
    const res = await api.put(`/api/keycloak/${courseId}/group/Editor/members/${UNKNOWN_USER_ID}`)
    expect(res.status()).toBe(404)
    const body = (await res.json()) as { error: string }
    expect(body.error).toMatch(/user not found/i)
  })
})

test.describe('core API: keycloak course-staff changes need a course lecturer', () => {
  // course-editor is an Editor of fullCourse, student a participant of iPraktikum:
  // both are course members, and neither may change the staff.
  const cases = [
    { role: 'course-editor', courseId: SEEDED_COURSES.fullCourse.id },
    { role: 'student', courseId: SEEDED_COURSES.iPraktikum.id },
  ] as const

  for (const { role, courseId } of cases) {
    test(`PUT and DELETE members are forbidden for a ${role}`, async ({ apiAs }) => {
      const api = await apiAs(role)
      const memberUrl = `/api/keycloak/${courseId}/group/Editor/members/${UNKNOWN_USER_ID}`
      expect((await api.put(memberUrl)).status()).toBe(403)
      expect((await api.delete(memberUrl)).status()).toBe(403)
    })
  }
})

test.describe('core API: keycloak status', () => {
  // The probes call the Keycloak Admin API with the prompt-server service account.
  // A realm or Keycloak change that takes that access away shows up here first.
  test('GET /keycloak/status reports a usable service account', async ({ apiAs }) => {
    const api = await apiAs('admin')
    const res = await api.get('/api/keycloak/status')
    expect(res.ok()).toBeTruthy()
    expect(await res.json()).toEqual({
      healthy: true,
      capabilities: {
        'keycloak.login': true,
        'keycloak.read_users': true,
        'keycloak.read_groups': true,
      },
    })
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
    const admin = results.results.find((u) => u.username === 'admin')
    expect(admin?.universityLogin).toBe('ad12min')
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
