import { type Page, type Locator, expect } from '@playwright/test'

// The public applicant application form at /apply/:phaseId. With
// universityLoginAvailable=false the login card offers a single "Continue as
// ... student" button that leads straight to the form (no Keycloak).
export class ApplicationFormPage {
  readonly fileInput: Locator

  constructor(private readonly page: Page) {
    this.fileInput = page.locator('input[type="file"]')
  }

  async goto(phaseId: string) {
    await this.page.goto(`/apply/${phaseId}`)
  }

  async continueToForm() {
    await this.page.getByRole('button', { name: /continue as .* student/i }).click()
    await expect(
      this.page.getByRole('heading', { name: /course specific questions/i }),
    ).toBeVisible()
  }

  // The FileUpload widget uploads immediately on selection (presign -> PUT ->
  // complete), so this is the whole upload action; no form submit needed.
  async uploadFile(name: string, buffer: Buffer, mimeType = 'text/plain') {
    await this.fileInput.setInputFiles({ name, mimeType, buffer })
  }

  async expectFileListed(name: string) {
    await expect(this.page.getByText('Uploaded file:')).toBeVisible()
    await expect(this.page.getByText(name)).toBeVisible()
  }
}
