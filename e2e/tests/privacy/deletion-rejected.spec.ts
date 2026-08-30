import type { APIRequestContext } from '@playwright/test'
import { SEEDED_PHASE_STUDENTS } from '../../src/data/constants'
import { apiContextFor } from '../../src/fixtures/api'
import { authFile, expect, test } from '../../src/fixtures/auth'
import { AdminPrivacyPage } from '../../src/pages/AdminPrivacyPage'
import { PrivacyPage } from '../../src/pages/PrivacyPage'
import { isDeletionTerminal, latestDeletion, PRIVACY_TEST_TIMEOUT_MS } from './helpers'

// The requester column renders the student's name, not their email.
const SUBJECT_NAME = `${SEEDED_PHASE_STUDENTS.student2.firstName} ${SEEDED_PHASE_STUDENTS.student2.lastName}`

// Rejection is the one review outcome that destroys nothing, so it runs against a
// shared seeded student. A rejected request is an end state, so the subject may
// request again and a retry starts from a clean slate.
test.describe('privacy: an admin rejects a deletion request', () => {
  test.use({ role: 'student2' })

  let subject: APIRequestContext
  let admin: APIRequestContext

  test.beforeAll(async () => {
    subject = await apiContextFor('student2')
    admin = await apiContextFor('admin')

    // A request left open by an interrupted run would make a new one 409.
    const existing = await latestDeletion(subject)
    if (existing.status === 'exists' && !isDeletionTerminal(existing.request.status)) {
      await admin.post(`/api/privacy/admin/data-deletions/${existing.request.id}`, {
        data: { decision: 'reject', note: 'cleaning up a previous e2e run' },
      })
    }
  })

  test.afterAll(async () => {
    await subject.dispose()
    await admin.dispose()
  })

  test('the subject sees the rejection, distinct from a failure', async ({ page, browser }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    const privacy = new PrivacyPage(page)
    await privacy.gotoDeletion()
    await privacy.requestDeletion()
    await privacy.expectBannerState('pending_approval')

    const adminContext = await browser.newContext({ storageState: authFile('admin') })
    try {
      const adminConsole = new AdminPrivacyPage(await adminContext.newPage())
      await adminConsole.goto()
      await adminConsole.expectLoaded()
      await adminConsole.openDeletionTab()

      await expect(adminConsole.rowFor(SUBJECT_NAME)).toContainText('Pending approval', {
        timeout: 15_000,
      })
      await adminConsole.openReview(SUBJECT_NAME)
      await adminConsole.reject()
    } finally {
      await adminContext.close()
    }

    await privacy.gotoDeletion()
    await privacy.expectBannerState('rejected', 30_000)
    await expect(privacy.statusBanner).toContainText('Deletion rejected')
    await expect(privacy.statusBanner).not.toContainText('Deletion failed')

    // A rejected request is terminal, so the subject may ask again.
    await expect(privacy.requestDeletionButton).toBeVisible()
  })
})
