import { type Locator, type Page, expect } from '@playwright/test'

export interface NewOrg {
  name: string
  slug?: string
  parent?: string
}

export class OrgsPage {
  constructor(private readonly page: Page) {}

  async goto() {
    await this.page.goto('/management/admin/orgs')
    await expect(this.page.getByRole('heading', { level: 1, name: 'Orgs' })).toBeVisible({
      timeout: 15_000,
    })
  }

  row(slug: string): Locator {
    return this.page
      .locator('table tbody tr')
      .filter({ has: this.page.getByRole('cell', { name: slug, exact: true }) })
  }

  async create({ name, slug, parent }: NewOrg) {
    await this.page.getByRole('button', { name: 'Add Org' }).click()
    const dialog = this.page.getByRole('dialog')
    await dialog.getByLabel('Name', { exact: true }).fill(name)
    if (slug !== undefined) {
      await dialog.getByLabel('Slug').fill(slug)
    }
    if (parent !== undefined) {
      await dialog.getByRole('combobox', { name: /^Parent Org/ }).click()
      await this.page.getByRole('option', { name: parent, exact: true }).click()
    }
    await dialog.getByRole('button', { name: 'Create' }).click()
    await expect(dialog).toBeHidden({ timeout: 15_000 })
  }

  async open(slug: string, name: string) {
    await this.row(slug).click()
    await expect(this.page.getByRole('heading', { level: 1, name })).toBeVisible({
      timeout: 15_000,
    })
  }
}

export class OrgDetailPage {
  constructor(private readonly page: Page) {}

  async expectLoaded(name: string) {
    await expect(this.page.getByRole('heading', { level: 1, name })).toBeVisible({
      timeout: 15_000,
    })
  }

  field(label: string): Locator {
    return this.page.getByLabel(label, { exact: true })
  }

  async saveDetails(values: Record<string, string>) {
    for (const [label, value] of Object.entries(values)) {
      await this.field(label).fill(value)
    }
    await this.clickAndAwaitPut(/\/api\/orgs\/[^/]+$/, 'Save')
  }

  async moveTo(parent: string) {
    await this.page.getByRole('combobox', { name: 'Parent Org' }).click()
    await this.page.getByRole('option', { name: parent, exact: true }).click()
    await this.clickAndAwaitPut(/\/api\/orgs\/[^/]+\/parent$/, 'Move')
  }

  private async clickAndAwaitPut(url: RegExp, button: string) {
    const [response] = await Promise.all([
      this.page.waitForResponse(
        (res) => res.request().method() === 'PUT' && url.test(new URL(res.url()).pathname),
      ),
      this.page.getByRole('button', { name: button, exact: true }).click(),
    ])
    expect(response.ok()).toBe(true)
  }

  childOrgLink(label: string): Locator {
    return this.page.getByRole('link', { name: label, exact: true })
  }

  async delete() {
    await this.page.getByRole('button', { name: 'Delete', exact: true }).click()
    await this.page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click()
  }
}
