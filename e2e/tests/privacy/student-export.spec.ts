import { request } from '@playwright/test'
import { authFile, expect, test } from '../../src/fixtures/auth'
import { SEEDED_PHASE_STUDENTS } from '../../src/data/constants'
import { ROLES } from '../../src/data/roles'
import { AdminPrivacyPage } from '../../src/pages/AdminPrivacyPage'
import { PrivacyPage } from '../../src/pages/PrivacyPage'
import {
  RoleApi,
  archiveExports,
  downloadUrlFor,
  getExport,
  isExportTerminal,
  latestExport,
  PRIVACY_TEST_TIMEOUT_MS,
  waitFor,
} from './helpers'

const SUBJECT_EMAIL = ROLES.student.email
const SUBJECT_NAME = `${SEEDED_PHASE_STUDENTS.student.firstName} ${SEEDED_PHASE_STUDENTS.student.lastName}`

// Only the Core document can succeed on this stack. The phase modules reject the
// presigned upload URL because prompt-sdk requires HTTPS for a non-localhost host
// and the e2e S3 endpoint is plain HTTP (http://seaweedfs-s3:8333), so the export
// finishes as "completed with issues". The journey under test - request, async
// completion, signed download, archive - is unaffected.
const CORE_DOCUMENT = 'Core'

test.describe('privacy: student data export', () => {
  test.use({ role: 'student' })

  const student = new RoleApi('student')
  const admin = new RoleApi('admin')

  test.beforeAll(async () => {
    // An export is rate limited for 30 days, so a previous run or a failed
    // teardown would otherwise make this spec unrunnable for the whole stack.
    await archiveExports(admin, SUBJECT_EMAIL)
  })

  test.afterAll(async () => {
    await archiveExports(admin, SUBJECT_EMAIL)
    await student.dispose()
    await admin.dispose()
  })

  test('a student requests an export, watches it complete and downloads it', async ({ page }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    const privacy = new PrivacyPage(page)
    await privacy.gotoOverview()
    await expect(privacy.overviewCards).toHaveCount(2)

    await privacy.gotoExport()
    await privacy.requestExport()

    // The banner appears as soon as the request exists; the work runs detached.
    await expect(privacy.statusBanner).toBeVisible({ timeout: 15_000 })

    const latest = await waitFor({
      describe: 'the data export',
      poll: () => latestExport(student),
      done: (value) => value.status === 'exists' && isExportTerminal(value.export.status),
    })
    if (latest.status !== 'exists') throw new Error('the export disappeared while polling')

    // 'partial' is the banner's own state for "completed with issues".
    await expect(privacy.statusBanner).toHaveAttribute('data-state', /^(success|partial)$/, {
      timeout: 30_000,
    })

    const produced = await getExport(student, latest.export.id)
    const coreDoc = produced.documents.find((doc) => doc.source_name === CORE_DOCUMENT)
    expect(coreDoc, `no Core document in ${JSON.stringify(produced.documents)}`).toBeTruthy()
    expect(coreDoc!.status).toBe('complete')

    // The button navigates to a presigned URL rather than emitting a download
    // event, and that navigation can outrun reading the response body, so assert
    // the request the page issues and resolve the URL through the API instead.
    const issued = page.waitForRequest(
      (req) => req.url().includes(`/docs/${coreDoc!.id}/download-url`),
      { timeout: 30_000 },
    )
    await privacy.documentDownloadButton(CORE_DOCUMENT).click()
    await issued

    const downloadUrl = await downloadUrlFor(student, latest.export.id, coreDoc!.id)
    const anonymous = await request.newContext()
    try {
      const file = await anonymous.get(downloadUrl)
      expect(file.status(), file.statusText()).toBe(200)
      expect(file.headers()['content-type']).toContain('zip')
      expect((await file.body()).length).toBeGreaterThan(0)
    } finally {
      await anonymous.dispose()
    }
  })

  test('an admin archives the export and frees the rate limit', async ({ browser }) => {
    test.setTimeout(PRIVACY_TEST_TIMEOUT_MS)

    const adminContext = await browser.newContext({ storageState: authFile('admin') })
    try {
      const adminConsole = new AdminPrivacyPage(await adminContext.newPage())
      await adminConsole.goto()
      await adminConsole.expectLoaded()
      await adminConsole.openExportTab()

      await expect(adminConsole.liveExportRowFor(SUBJECT_NAME).first()).toBeVisible({
        timeout: 15_000,
      })
      await adminConsole.archiveExportAndResetRateLimit(SUBJECT_NAME)

      await waitFor({
        describe: 'the export archive',
        poll: () => latestExport(student),
        done: (value) => value.status !== 'exists' || value.export.status === 'archived',
        timeoutMs: 30_000,
      })
    } finally {
      await adminContext.close()
    }
  })
})
