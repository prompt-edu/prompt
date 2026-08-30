import { PRIVACY_SUBJECTS } from '../../src/data/constants'
import { expect, test } from '../../src/fixtures/auth'
import { StudentsPage } from '../../src/pages/StudentsPage'
import { PRIVACY_TEST_TIMEOUT_MS, RoleApi, waitFor } from './helpers'

const SUBJECT = PRIVACY_SUBJECTS.inactive

async function studentExists(admin: RoleApi): Promise<boolean> {
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

  const admin = new RoleApi('admin')

  test.afterAll(async () => {
    await admin.dispose()
  })

  test('an admin filters for inactivity and deletes the student', async ({ page }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    if (!(await studentExists(admin))) {
      // A previous attempt already deleted the subject. Assert the deletion that
      // did it actually succeeded, rather than that the subject is missing - which
      // a half-finished run would also satisfy.
      const requests = await admin.get('/api/privacy/admin/data-deletions')
      expect(requests.ok(), await requests.text()).toBeTruthy()
      const all = (await requests.json()) as { status: string; student_id: string | null }[]
      expect(
        all.some((request) => request.status === 'succeeded'),
        'the subject is gone but no deletion succeeded',
      ).toBe(true)
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

    // "Deletion failed" and the in-progress "N of M done" line both sit next to a
    // Close button, so the dialog's own success wording is the only safe assertion.
    await expect(students.deletionDialog).toContainText('Deletion complete', {
      timeout: 60_000,
    })
  })
})
