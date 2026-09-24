import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId, the Infrastructure Setup remote
// (Module Federation) rendered inside the core shell. The phase root is the
// student page, which renders the phase name as an <h1> for staff and students
// alike; the lecturer pages are Participants, Configuration and Provisioning.
export class InfrastructureSetupPage {
  readonly title: Locator

  constructor(private readonly page: Page) {
    this.title = this.page.getByRole('heading', { name: 'Infrastructure Setup', level: 1 })
  }

  async goto(courseId: string, phaseId: string, subPath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subPath}`)
  }

  async expectLoaded() {
    await expect(this.title).toBeVisible({ timeout: 15_000 })
  }

  // Remote toasts render through core's single <Toaster />, which only works while
  // host and remote share one copy of the UI library.
  async expectToast(title: string | RegExp) {
    await expect(this.page.getByText(title).first()).toBeVisible({ timeout: 10_000 })
  }

  dialog(): Locator {
    return this.page.getByRole('dialog')
  }

  // ── Configuration ────────────────────────────────────────────────────────

  section(id: 'semester-tag' | 'providers' | 'resources'): Locator {
    return this.page.locator(`section#${id}`)
  }

  async saveSemesterTag(tag: string) {
    const section = this.section('semester-tag')
    await section.getByLabel('Semester tag').fill(tag)
    await section.getByRole('button', { name: 'Save' }).click()
  }

  async addKeycloakProvider(provider: {
    url: string
    realm: string
    clientId: string
    clientSecret: string
  }) {
    await this.section('providers').getByRole('button', { name: 'Add provider' }).click()
    const dialog = this.dialog()
    await dialog.locator('#providerType').click()
    await this.page.getByRole('option', { name: 'keycloak' }).click()
    await dialog.locator('#keycloak_url').fill(provider.url)
    await dialog.locator('#realm').fill(provider.realm)
    await dialog.locator('#client_id').fill(provider.clientId)
    await dialog.locator('#client_secret').fill(provider.clientSecret)
    await dialog.getByRole('button', { name: 'Save' }).click()
  }

  // The only configured provider is preselected, and so is its only resource kind.
  async addPerStudentResource(nameTemplate: string) {
    await this.section('resources').getByRole('button', { name: 'Add resource' }).click()
    const dialog = this.dialog()
    await dialog.getByLabel('per student').click()
    await dialog.locator('#nameTemplate').fill(nameTemplate)
    await dialog.getByRole('button', { name: 'Save' }).click()
  }

  // ── Provisioning ─────────────────────────────────────────────────────────

  provisionButton(): Locator {
    return this.page.getByRole('button', { name: 'Provision resources' })
  }

  statusFilter(label: string): Locator {
    return this.page.getByRole('button', { name: label, exact: true })
  }
}
