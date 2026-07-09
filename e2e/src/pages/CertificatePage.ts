import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId — the Certificate remote (Module
// Federation) rendered inside the core shell. Staff configure the template and
// release date and see the participants download table; students see their
// certificate status and self-download once released.
export class CertificatePage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string, subpath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subpath}`)
  }

  // ── Student overview (phase root) ────────────────────────────────────────

  async expectOverviewLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Course Certificate' })).toBeVisible({
      timeout: 15_000,
    })
  }

  downloadButton(): Locator {
    return this.page.getByRole('button', { name: 'Download Certificate' })
  }

  async expectNotAvailable(message: string | RegExp) {
    await expect(this.page.getByText(message)).toBeVisible({ timeout: 15_000 })
    await expect(this.downloadButton()).toBeHidden()
  }

  // Clicks the student self-download button and returns the captured download.
  async downloadOwnCertificate() {
    const downloadPromise = this.page.waitForEvent('download')
    await this.downloadButton().click()
    return downloadPromise
  }

  async expectLastDownloaded() {
    await expect(this.page.getByText(/Last downloaded:/)).toBeVisible({ timeout: 15_000 })
  }

  // ── Settings ─────────────────────────────────────────────────────────────

  async expectSettingsLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Certificate Settings' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async pasteTemplate(content: string) {
    await this.page.getByPlaceholder('Paste your Typst template content here...').fill(content)
  }

  async saveTemplate() {
    await this.page.getByRole('button', { name: 'Save Template' }).click()
    // The "Configured" badge appends "· <date> by <user>", so match a substring.
    await expect(this.page.getByText(/Configured/)).toBeVisible({ timeout: 15_000 })
  }

  // Generates a preview PDF from the saved template. Awaiting the preview
  // response asserts the template actually compiled server-side (200); a broken
  // template would return 422 and fail here.
  async testCertificate() {
    const responsePromise = this.page.waitForResponse(
      (res) => res.url().includes('/certificate/preview') && res.request().method() === 'GET',
    )
    await this.page.getByRole('button', { name: 'Test Certificate' }).click()
    const response = await responsePromise
    expect(response.ok()).toBeTruthy()
  }

  // Awaits the release-date response so the release is confirmed persisted
  // before the caller navigates away (which would otherwise abort the request).
  async releaseNow() {
    const responsePromise = this.page.waitForResponse(
      (res) => res.url().includes('/config/release-date') && res.request().method() === 'PUT',
    )
    await this.page.getByRole('button', { name: 'Release Now' }).click()
    const response = await responsePromise
    expect(response.ok()).toBeTruthy()
  }

  // ── Participants ─────────────────────────────────────────────────────────

  async expectParticipantsLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Certificate Participants' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async expectParticipantRow(fullName: string, downloadStatus: string) {
    const row = this.page.getByRole('row', { name: new RegExp(fullName) })
    await expect(row).toBeVisible()
    await expect(row).toContainText(downloadStatus)
  }
}
