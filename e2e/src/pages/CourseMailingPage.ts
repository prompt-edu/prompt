import { type Locator, type Page, expect } from '@playwright/test'

interface CampaignInput {
  name: string
  phaseName: string
  status: string
  subject: string
  body: string
  replyToEmail: string
}

// /management/course/:id/mailing: campaign overview + composer.
export class CourseMailingPage {
  readonly newMailButton: Locator

  constructor(
    private readonly page: Page,
    private readonly courseId: string,
  ) {
    this.newMailButton = page.getByRole('button', { name: 'New mail' })
  }

  async goto() {
    await this.page.goto(`/management/course/${this.courseId}/mailing`)
  }

  async expectLoaded() {
    await expect(this.newMailButton).toBeVisible({ timeout: 15_000 })
  }

  async startNewCampaign() {
    await this.newMailButton.click()
    await expect(this.page.getByRole('heading', { name: 'New Campaign' })).toBeVisible()
  }

  async fillDetails(input: CampaignInput) {
    await this.page.getByLabel('Name', { exact: true }).fill(input.name)

    // Course phase (Radix Select → combobox trigger).
    await this.page.getByRole('combobox').click()
    await this.page.getByRole('option', { name: input.phaseName }).click()

    // Student statuses (MultiSelect popover).
    await this.page.getByText('Select statuses').click()
    await this.page.getByRole('option', { name: input.status }).click()
    await this.page.keyboard.press('Escape')

    // Subject + body (shared EmailTemplateEditor).
    await this.page.locator('input[name="subject"]').fill(input.subject)
    const editor = this.page.locator('[contenteditable="true"]').first()
    await editor.click()
    await editor.fill(input.body)

    // Per-campaign reply-to override (avoids depending on course mailing config).
    await this.page.getByLabel('Reply-to override (optional)').fill(input.replyToEmail)
  }

  async save() {
    await this.page.getByRole('button', { name: 'Save', exact: true }).click()
    await expect(this.page.getByText(/Draft created|Campaign saved/)).toBeVisible()
  }

  async showRecipients() {
    await this.page.getByRole('button', { name: /Show recipients/ }).click()
  }

  async sendWithConfirmation() {
    await this.page.getByRole('button', { name: 'Send', exact: true }).click()
    const dialog = this.page.getByRole('dialog')
    await expect(dialog).toBeVisible()
    await dialog.getByRole('button', { name: /Send to \d+/ }).click()
  }
}
