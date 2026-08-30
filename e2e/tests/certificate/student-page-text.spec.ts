import { authFile, expect, test } from '../../src/fixtures/auth'
import { CERTIFICATE_FIXTURE_PHASES, SEEDED_COURSES } from '../../src/data/constants'
import { CertificatePage } from '../../src/pages/CertificatePage'
import { E2E_TEMPLATE, putTemplate, resetCertificatePhase, setStudentPageText } from './helpers'

const PHASE_ID = CERTIFICATE_FIXTURE_PHASES.studentPageText
const COURSE_ID = SEEDED_COURSES.fullCourse.id
const MESSAGE = 'Add this certificate to your LinkedIn profile.'

// Instructor HTML reaches a student's browser, so the sanitizer is exercised
// with markup an attacker would use rather than only with a benign paragraph.
const XSS_MARKER = '__certificateXss'
const HOSTILE_HTML =
  `<p>${MESSAGE}</p>` +
  '<p><a href="https://example.com/guide">Certificate guide</a></p>' +
  `<script>window.${XSS_MARKER}=1</script>` +
  `<img src="x" onerror="window.${XSS_MARKER}=1">` +
  `<p><a href="javascript:window.${XSS_MARKER}=1">do not click</a></p>`

test.describe('certificate: instructor text on the student download page', () => {
  test.afterAll(async () => {
    await resetCertificatePhase(PHASE_ID)
  })

  test.describe('as a student', () => {
    test.use({ role: 'student' })

    test.beforeAll(async () => {
      await setStudentPageText(PHASE_ID, HOSTILE_HTML)
    })

    // Deliberately makes no claim about whether a template is configured: the
    // lecturer case below uploads one and there is no API to remove it, so a CI
    // retry would find the phase in the other state.
    test('the instructor text renders, sanitized', async ({ page }) => {
      const phase = new CertificatePage(page)
      await phase.goto(COURSE_ID, PHASE_ID)
      await phase.expectOverviewLoaded()

      await expect(phase.studentPageText).toContainText(MESSAGE, { timeout: 15_000 })

      const link = phase.studentPageText.getByRole('link', { name: 'Certificate guide' })
      await expect(link).toHaveAttribute('target', '_blank')
      await expect(link).toHaveAttribute('rel', 'noopener noreferrer')

      await expect(phase.studentPageText.locator('script')).toHaveCount(0)
      await expect(phase.studentPageText.locator('[onerror]')).toHaveCount(0)
      await expect(phase.studentPageText.locator('a[href^="javascript:"]')).toHaveCount(0)
      expect(await phase.readGlobal(XSS_MARKER)).toBeUndefined()
    })

    test('an unset text leaves the page as it was', async ({ page }) => {
      await setStudentPageText(PHASE_ID, null)

      const phase = new CertificatePage(page)
      await phase.goto(COURSE_ID, PHASE_ID)
      await phase.expectOverviewLoaded()
      await expect(phase.studentPageText).toHaveCount(0)

      await setStudentPageText(PHASE_ID, HOSTILE_HTML)
    })
  })

  test.describe('as a lecturer', () => {
    test.use({ role: 'lecturer' })

    test('a lecturer writes the text and the student sees it', async ({ page, browser }) => {
      const settings = new CertificatePage(page)
      await settings.goto(COURSE_ID, PHASE_ID, '/settings')
      await settings.expectSettingsLoaded()

      const written = 'Questions about your certificate? Write to the course staff.'
      await settings.setStudentPageText(written)

      // It survives a reload, so it was persisted and not only held in state.
      await settings.goto(COURSE_ID, PHASE_ID, '/settings')
      await expect(settings.studentPageTextEditor).toContainText(written, { timeout: 15_000 })

      // A released certificate must carry the text too, not only the "not
      // available" state.
      await putTemplate(PHASE_ID, E2E_TEMPLATE)

      const studentContext = await browser.newContext({ storageState: authFile('student') })
      try {
        const studentView = new CertificatePage(await studentContext.newPage())
        await studentView.goto(COURSE_ID, PHASE_ID)
        await studentView.expectOverviewLoaded()
        await expect(studentView.downloadButton()).toBeVisible({ timeout: 15_000 })
        await expect(studentView.studentPageText).toContainText(written)
      } finally {
        await studentContext.close()
      }
    })
  })
})
