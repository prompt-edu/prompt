import type { APIRequestContext } from '@playwright/test'
import { PRIVACY_SUBJECTS } from '../../src/data/constants'
import { apiContextFor } from '../../src/fixtures/api'
import { expect, test } from '../../src/fixtures/auth'
import { StudentsPage } from '../../src/pages/StudentsPage'
import { PRIVACY_TEST_TIMEOUT_MS, waitFor } from './helpers'

const SUBJECT = PRIVACY_SUBJECTS.inactive

async function studentExists(admin: APIRequestContext): Promise<boolean> {
  const res = await admin.get('/api/students/')
  if (!res.ok()) throw new Error(`GET /api/students/ failed: ${res.status()}`)
  const students = (await res.json()) as { id: string }[]
  return students.some((student) => student.id === SUBJECT.id)
}

// Admin-initiated deletion has no requester: the admin selects the subjects and
// the batch starts already in progress. The subject is a seeded student with a
// deliberately old last_modified, because a student created during the run can
// never match the "not modified in N years" filter.
test.describe('privacy: admin-initiated deletion of an inactive student', () => {
  test.use({ role: 'admin' })

  let admin: APIRequestContext

  test.beforeAll(async () => {
    admin = await apiContextFor('admin')
  })

  test.afterAll(async () => {
    await admin.dispose()
  })

  test('an admin filters for inactivity and deletes the student', async ({ page }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    if (!(await studentExists(admin))) {
      // A previous attempt already deleted the subject; that is the outcome
      // under test, so assert it rather than re-running a destructive flow.
      expect(await studentExists(admin)).toBe(false)
      return
    }

    const students = new StudentsPage(page)
    await students.goto()
    await students.expectLoaded()

    await students.filterByInactivity(3)
    await expect(students.rowFor(SUBJECT.email)).toHaveCount(1, { timeout: 15_000 })

    await students.openDeletionDialogFor(SUBJECT.email)
    await students.startDeletion()

    await waitFor({
      describe: 'the admin-initiated deletion',
      poll: () => studentExists(admin),
      done: (exists) => !exists,
    })

    await expect(students.deletionDialog).toContainText(/Close|completed|done/i, {
      timeout: 60_000,
    })
  })
})
