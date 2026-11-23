import { type APIRequestContext, expect } from '@playwright/test'
import { SEEDED_ORG } from '../../src/data/constants'
import type { Role } from '../../src/data/roles'
import { apiContextFor } from '../../src/fixtures/api'
import { test } from '../../src/fixtures/auth'
import { CopyCoursePage } from '../../src/pages/CopyCoursePage'
import { CourseCreationPage } from '../../src/pages/CourseCreationPage'
import { CourseSettingsPage } from '../../src/pages/CourseSettingsPage'
import { OrgDetailPage } from '../../src/pages/OrgsPage'
import { createCourseViaApi, deleteCourseViaApi } from '../courses/helpers'

// A course and a template of the seeded org, created by a PROMPT admin before the tests. The Admins
// of an org are Lecturer-equivalent on all of its courses; its Members create courses in it, also
// from its templates.

interface CourseOrg {
  orgID?: string
}

const asRole = async <T>(role: Role, run: (api: APIRequestContext) => Promise<T>): Promise<T> => {
  const api = await apiContextFor(role)
  try {
    return await run(api)
  } finally {
    await api.dispose()
  }
}

const unique = (prefix: string) => `${prefix}${Date.now()}${Math.floor(Math.random() * 1000)}`

const orgOf = (courseId: string): Promise<string | undefined> =>
  asRole('admin', async (admin) => {
    const res = await admin.get(`/api/courses/${courseId}`)
    expect(res.status()).toBe(200)
    return ((await res.json()) as CourseOrg).orgID
  })

const created: string[] = []
let orgCourse = { id: '', name: '' }
let orgTemplate = { id: '', name: '' }

test.beforeAll(async () => {
  await asRole('admin', async (admin) => {
    const courseName = unique('OrgCourse')
    const templateName = unique('OrgTemplate')
    orgCourse = {
      id: await createCourseViaApi(admin, {
        name: courseName,
        semesterTag: 'orgtest',
        orgID: SEEDED_ORG.id,
      }),
      name: courseName,
    }
    orgTemplate = {
      id: await createCourseViaApi(admin, {
        name: templateName,
        semesterTag: 'orgtest',
        template: true,
        orgID: SEEDED_ORG.id,
      }),
      name: templateName,
    }
  })
  created.push(orgCourse.id, orgTemplate.id)
})

test.afterAll(async () => {
  await asRole('admin', async (admin) => {
    for (const id of created) {
      await deleteCourseViaApi(admin, id)
    }
  })
})

test.describe('as the org admin', () => {
  test.use({ role: 'org-admin' })

  test('manages the courses of the org', async ({ page }) => {
    await page.goto('/management/courses')
    await expect(page.getByText(orgCourse.name).first()).toBeVisible({ timeout: 15_000 })

    const settings = new CourseSettingsPage(page, orgCourse.id)
    await settings.goto()
    await settings.expectLoaded()
    await settings.editShortDescription(`Edited by the org admin ${Date.now()}`)
    await settings.expectOrg(SEEDED_ORG.name)

    await page.goto(`/management/org/${SEEDED_ORG.id}`)
    await new OrgDetailPage(page).expectLoaded(SEEDED_ORG.name)
    await expect(page.getByText(orgCourse.name).first()).toBeVisible()
  })

  test('copies a course of the org only into an org', async () => {
    await asRole('org-admin', async (api) => {
      const copy = (orgID?: string) =>
        api.post(`/api/courses/${orgCourse.id}/copy`, {
          data: {
            name: unique('OrgCopy'),
            semesterTag: 'orgtest',
            startDate: '2026-10-01',
            endDate: '2027-02-01',
            orgID,
          },
        })

      expect((await copy()).status()).toBe(403)
      const res = await copy(SEEDED_ORG.id)
      expect(res.status()).toBe(201)
      created.push(((await res.json()) as { id: string }).id)
    })
  })
})

test.describe('as the org member', () => {
  test.use({ role: 'org-member' })

  test('cannot manage the courses of the org', async () => {
    await asRole('org-member', async (api) => {
      expect((await api.get(`/api/courses/${orgCourse.id}`)).status()).toBe(403)
      expect((await api.put(`/api/courses/${orgCourse.id}/org`, { data: { orgID: null } })).status()).toBe(
        403,
      )
    })
  })

  test('creates a course in the org', async ({ page }) => {
    const courseId = await new CourseCreationPage(page).create({
      name: unique('MemberCourse'),
      semesterTag: 'orgtest',
      org: SEEDED_ORG.name,
    })
    created.push(courseId)

    expect(await orgOf(courseId)).toBe(SEEDED_ORG.id)
  })

  test('creates a course from a template of the org', async ({ page }) => {
    const courseId = await new CopyCoursePage(page).copyFromTemplate({
      templateName: orgTemplate.name,
      name: unique('FromTemplate'),
      semesterTag: 'orgtest',
      org: SEEDED_ORG.name,
    })
    created.push(courseId)

    expect(await orgOf(courseId)).toBe(SEEDED_ORG.id)
  })
})

test.describe('as a PROMPT admin', () => {
  test.use({ role: 'admin' })

  test('assigns a course to an org and removes it again', async ({ page }) => {
    const courseId = await asRole('admin', (admin) =>
      createCourseViaApi(admin, { name: unique('Ungrouped'), semesterTag: 'orgtest' }),
    )
    created.push(courseId)

    const settings = new CourseSettingsPage(page, courseId)
    await settings.goto()
    await settings.expectLoaded()

    await settings.setOrg(SEEDED_ORG.name)
    await expect.poll(() => orgOf(courseId)).toBe(SEEDED_ORG.id)

    await settings.setOrg('No org')
    await expect.poll(() => orgOf(courseId)).toBeUndefined()
  })
})
