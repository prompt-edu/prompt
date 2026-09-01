import { PRIVACY_SUBJECTS } from '../../src/data/constants'
import { authFile, expect, test } from '../../src/fixtures/auth'
import { AdminPrivacyPage } from '../../src/pages/AdminPrivacyPage'
import { PrivacyPage } from '../../src/pages/PrivacyPage'
import {
  RoleApi,
  adminDeletion,
  isDeletionTerminal,
  latestDeletion,
  PRIVACY_TEST_TIMEOUT_MS,
  waitFor,
} from './helpers'

const SUBJECT = PRIVACY_SUBJECTS.deletionApproval

async function studentExists(admin: RoleApi): Promise<boolean> {
  const res = await admin.get('/api/students/')
  if (!res.ok()) throw new Error(`GET /api/students/ failed: ${res.status()}`)
  const students = (await res.json()) as { id: string }[]
  return students.some((student) => student.id === SUBJECT.id)
}

// The only journey that genuinely deletes its subject, so it owns a Keycloak user
// and a student row that nothing else touches. It is written to be state-aware
// rather than skippable: a retry after the destructive step asserts the real end
// state instead of quietly passing.
test.describe('privacy: student deletion request approved by an admin', () => {
  test.use({ role: 'privacy-student' })

  const subject = new RoleApi('privacy-student')
  const admin = new RoleApi('admin')

  test.afterAll(async () => {
    await subject.dispose()
    await admin.dispose()
  })

  test('the request is approved and the subject is removed', async ({ page, browser }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    let existing = await latestDeletion(subject)

    if (existing.status === 'exists' && isDeletionTerminal(existing.request.status)) {
      // A previous attempt already ran it. Only a successful deletion may pass:
      // a failed one can leave the subject half-deleted.
      expect(existing.request.status, JSON.stringify(existing.request)).toBe('succeeded')
      expect(await studentExists(admin)).toBe(false)
      return
    }

    if (existing.status === 'ready') {
      const privacy = new PrivacyPage(page)
      await privacy.gotoDeletion()
      await privacy.requestDeletion()
      await privacy.expectBannerState('pending_approval')

      existing = await latestDeletion(subject)
    }

    if (existing.status !== 'exists') throw new Error('the deletion request was not created')
    const requestID = existing.request.id

    if (existing.request.status === 'pending_approval') {
      const adminContext = await browser.newContext({ storageState: authFile('admin') })
      try {
        const adminConsole = new AdminPrivacyPage(await adminContext.newPage())
        await adminConsole.goto()
        await adminConsole.expectLoaded()
        await adminConsole.openDeletionTab()

        const row = adminConsole.pendingRowFor(SUBJECT.name)
        await expect(row).toHaveCount(1, { timeout: 15_000 })

        // The server sorts pending requests first. Which pending request leads is
        // not this spec's business: another spec may legitimately have a newer one.
        await expect(adminConsole.rows.first()).toContainText('Pending approval')

        await adminConsole.openReview(SUBJECT.name)
        await adminConsole.approve()
      } finally {
        await adminContext.close()
      }
    }

    const finished = await waitFor({
      describe: 'the approved deletion',
      poll: () => adminDeletion(admin, requestID),
      done: (value) => isDeletionTerminal(value.status),
      failed: (value) => value.status === 'failed',
    })
    expect(finished.status).toBe('succeeded')

    expect(await studentExists(admin), 'the subject must be gone from core').toBe(false)
  })
})
