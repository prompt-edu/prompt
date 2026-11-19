import { randomUUID } from 'node:crypto'
import { type APIRequestContext, expect, type Page } from '@playwright/test'
import { SEEDED_ORG } from '../../src/data/constants'
import type { Role } from '../../src/data/roles'
import { apiContextFor } from '../../src/fixtures/api'
import { test } from '../../src/fixtures/auth'
import { OrgDetailPage } from '../../src/pages/OrgsPage'

// The account the tests add to the seeded org. No other spec uses it, and every test
// that changes the org's membership lives in this file, so they never run concurrently.
const ADDED_USER = { username: 'student-passkey', universityLogin: 'no43tum' }

const MEMBERS_PATH = `/api/orgs/${SEEDED_ORG.id}/members`

interface UserSearchResults {
  results: { keycloakUserID: string; username: string }[]
}

interface OrgMember {
  keycloakUserID: string
  username: string
  universityLogin: string
  role: string
}

interface OwnOrg {
  id: string
  role: string
}

const asRole = async <T>(role: Role, run: (api: APIRequestContext) => Promise<T>): Promise<T> => {
  const api = await apiContextFor(role)
  try {
    return await run(api)
  } finally {
    await api.dispose()
  }
}

const membersAs = (role: Role): Promise<OrgMember[]> =>
  asRole(role, async (api) => (await (await api.get(MEMBERS_PATH)).json()) as OrgMember[])

// The Keycloak ID of a user, found through the search that the Add Member dialog uses
const keycloakIDOf = (role: Role, username: string): Promise<string> =>
  asRole(role, async (api) => {
    const res = await api.get(`/api/keycloak/users/search?q=${encodeURIComponent(username)}`)
    expect(res.status()).toBe(200)
    const { results } = (await res.json()) as UserSearchResults
    const user = results.find((result) => result.username === username)
    expect(user, username).toBeDefined()
    return user?.keycloakUserID ?? ''
  })

const removeAddedUser = () =>
  asRole('admin', async (admin) => {
    const members = (await (await admin.get(MEMBERS_PATH)).json()) as OrgMember[]
    for (const member of members.filter((m) => m.username === ADDED_USER.username)) {
      await admin.delete(`${MEMBERS_PATH}/${member.keycloakUserID}`)
    }
  })

// Through the API, so the seeded details come back even when the page that changed them broke
const restoreSeededDetails = () =>
  asRole('admin', async (admin) => {
    const res = await admin.put(`/api/orgs/${SEEDED_ORG.id}`, { data: SEEDED_ORG.details })
    expect(res.ok()).toBe(true)
  })

const openOwnOrgFromSidebar = async (page: Page) => {
  await page.goto('/management/courses')
  await expect(page.getByText('My Org', { exact: true })).toBeVisible({ timeout: 15_000 })
  await page.getByRole('button', { name: SEEDED_ORG.name, exact: true }).click()
  const detail = new OrgDetailPage(page)
  await detail.expectLoaded(SEEDED_ORG.name)
  await expect(page).toHaveURL(new RegExp(`/management/org/${SEEDED_ORG.id}$`))
  return detail
}

test.describe('org membership', () => {
  test.afterEach(removeAddedUser)

  test.describe('as the org admin', () => {
    test.use({ role: 'org-admin' })

    test('reaches the org through the sidebar and edits its details', async ({ page }) => {
      const detail = await openOwnOrgFromSidebar(page)
      await expect(page.getByText('Danger Zone')).toBeHidden()
      await expect(page.getByRole('combobox', { name: 'Parent Org' })).toBeHidden()

      // A value no earlier run left behind, so Save is enabled even after a failed restore
      const newSchool = `School of E2E Testing ${Date.now()}`
      const school = detail.field('School (optional)')
      try {
        await detail.saveDetails({ 'School (optional)': newSchool })
        await page.reload()
        await detail.expectLoaded(SEEDED_ORG.name)
        await expect(school).toHaveValue(newSchool)
      } finally {
        await restoreSeededDetails()
      }
    })

    test('adds a member, changes their role, and removes them', async ({ page }) => {
      const detail = await openOwnOrgFromSidebar(page)
      await expect(detail.memberRow('or11adm')).toBeVisible()
      await expect(detail.memberRow('or12mem')).toBeVisible()

      await detail.addMember(ADDED_USER.username, ADDED_USER.username, 'Member')
      const row = detail.memberRow(ADDED_USER.universityLogin)
      await expect(row).toBeVisible()
      await expect(detail.memberRoleSelect(ADDED_USER.username)).toHaveText('Member')

      await detail.setMemberRole(ADDED_USER.username, 'Admin')
      await expect(detail.memberRoleSelect(ADDED_USER.username)).toHaveText('Admin')

      await detail.removeMember(ADDED_USER.username)
      await expect(row).toBeHidden()
    })

    test('cannot change their own membership', async ({ page }) => {
      const detail = await openOwnOrgFromSidebar(page)

      await expect(detail.memberRoleSelect('org-admin')).toBeDisabled()
      await expect(detail.removeMemberButton('org-admin')).toBeDisabled()
      await expect(detail.memberRoleSelect('org-member')).toBeEnabled()
    })

    test('finds existing members in the search but cannot add them again', async ({ page }) => {
      const detail = await openOwnOrgFromSidebar(page)

      const dialog = await detail.searchUsersToAdd('org-member')

      await expect(dialog.getByText('Already in the org as Member')).toBeVisible()
      await expect(dialog.getByRole('button', { name: 'Add org-member', exact: true })).toBeDisabled()
    })
  })

  test.describe('as the org member', () => {
    test.use({ role: 'org-member' })

    test('sees the org read-only and no members', async ({ page }) => {
      const detail = await openOwnOrgFromSidebar(page)

      await expect(detail.field('Name')).toBeDisabled()
      await expect(page.getByRole('button', { name: 'Save', exact: true })).toBeHidden()
      await expect(detail.membersHeading()).toBeHidden()
      await expect(page.getByText('Danger Zone')).toBeHidden()
    })
  })

  test.describe('as a user without org roles', () => {
    test.use({ role: 'lecturer' })

    test('has no org entry in the sidebar', async ({ page }) => {
      await page.goto('/management/courses')
      await expect(page.getByRole('heading', { level: 1, name: 'Courses' })).toBeVisible({
        timeout: 15_000,
      })

      await expect(page.getByText('My Org', { exact: true })).toBeHidden()
      await expect(page.getByRole('button', { name: SEEDED_ORG.name, exact: true })).toBeHidden()
    })
  })

  test.describe('through the API', () => {
    test('lists the seeded Admin and Member', async () => {
      const members = await membersAs('org-admin')

      expect(members.map(({ username, role }) => ({ username, role }))).toEqual(
        expect.arrayContaining([
          { username: 'org-admin', role: 'Admin' },
          { username: 'org-member', role: 'Member' },
        ]),
      )
    })

    test('an org admin gives a user one role at a time and removes them', async () => {
      const userID = await keycloakIDOf('org-admin', ADDED_USER.username)
      const rolesOfAddedUser = async () =>
        (await membersAs('org-admin'))
          .filter((m) => m.keycloakUserID === userID)
          .map((m) => m.role)

      await asRole('org-admin', async (orgAdmin) => {
        const added = await orgAdmin.post(`${MEMBERS_PATH}/${userID}`, { data: { role: 'Member' } })
        expect(added.status()).toBe(204)
        expect(await rolesOfAddedUser()).toEqual(['Member'])
        const listed = (await membersAs('org-admin')).find((m) => m.keycloakUserID === userID)
        expect(listed?.universityLogin).toBe(ADDED_USER.universityLogin)

        // Adding never changes the role of a member, so a stale pick cannot demote anyone
        const readded = await orgAdmin.post(`${MEMBERS_PATH}/${userID}`, { data: { role: 'Admin' } })
        expect(readded.status()).toBe(409)
        expect(await rolesOfAddedUser()).toEqual(['Member'])

        const promoted = await orgAdmin.put(`${MEMBERS_PATH}/${userID}`, { data: { role: 'Admin' } })
        expect(promoted.status()).toBe(204)
        expect(await rolesOfAddedUser()).toEqual(['Admin'])

        const removed = await orgAdmin.delete(`${MEMBERS_PATH}/${userID}`)
        expect(removed.status()).toBe(204)
        expect(await rolesOfAddedUser()).toEqual([])
      })
    })

    test('an unknown user is not found', async () => {
      await asRole('org-admin', async (orgAdmin) => {
        const res = await orgAdmin.post(`${MEMBERS_PATH}/${randomUUID()}`, {
          data: { role: 'Member' },
        })

        expect(res.status()).toBe(404)
      })
    })

    test('an org admin cannot change their own membership', async () => {
      const self = (await membersAs('org-admin')).find((m) => m.username === 'org-admin')
      expect(self).toBeDefined()

      await asRole('org-admin', async (orgAdmin) => {
        const demoted = await orgAdmin.put(`${MEMBERS_PATH}/${self?.keycloakUserID}`, {
          data: { role: 'Member' },
        })
        const left = await orgAdmin.delete(`${MEMBERS_PATH}/${self?.keycloakUserID}`)

        expect(demoted.status()).toBe(400)
        expect(left.status()).toBe(400)
      })
    })

    test('a PROMPT admin may add and remove themselves', async () => {
      const adminID = await keycloakIDOf('admin', 'admin')

      await asRole('admin', async (admin) => {
        try {
          const joined = await admin.post(`${MEMBERS_PATH}/${adminID}`, { data: { role: 'Admin' } })
          expect(joined.status()).toBe(204)
          expect((await membersAs('admin')).some((m) => m.keycloakUserID === adminID)).toBe(true)
        } finally {
          const left = await admin.delete(`${MEMBERS_PATH}/${adminID}`)
          expect(left.status()).toBe(204)
        }
      })
    })

    test('a member id that is not a UUID is rejected', async () => {
      await asRole('org-admin', async (orgAdmin) => {
        for (const userID of ['not-a-uuid', '00000000-0000-0000-0000-00000000000x']) {
          const removed = await orgAdmin.delete(`${MEMBERS_PATH}/${userID}`)
          const promoted = await orgAdmin.put(`${MEMBERS_PATH}/${userID}`, { data: { role: 'Admin' } })
          expect(removed.status(), userID).toBe(400)
          expect(promoted.status(), userID).toBe(400)
        }
      })
    })

    test('an org member can neither manage members nor search for users to add', async () => {
      await asRole('org-member', async (orgMember) => {
        const added = await orgMember.post(`${MEMBERS_PATH}/${randomUUID()}`, {
          data: { role: 'Member' },
        })
        const searched = await orgMember.get('/api/keycloak/users/search?q=student')

        expect(added.status()).toBe(403)
        expect(searched.status()).toBe(403)
      })
    })

    test('the own orgs follow the token roles', async () => {
      for (const [role, expected] of [
        ['org-admin', [{ id: SEEDED_ORG.id, role: 'Admin' }]],
        ['org-member', [{ id: SEEDED_ORG.id, role: 'Member' }]],
        ['admin', []],
        ['lecturer', []],
      ] as const) {
        const ownOrgs = await asRole(role, async (api) => {
          const res = await api.get('/api/orgs/self')
          expect(res.status(), role).toBe(200)
          return (await res.json()) as OwnOrg[]
        })

        expect(ownOrgs.map(({ id, role: orgRole }) => ({ id, role: orgRole })), role).toEqual(expected)
      }
    })
  })
})
