import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId — the Example remote (Module Federation)
// rendered inside the core shell. The example phase is a minimal placeholder,
// so its overview only shows a static "Example Component" card (the title is a
// shadcn CardTitle <div>, not a heading element, so match it by text).
export class ExamplePage {
  readonly title: Locator

  constructor(private readonly page: Page) {
    this.title = this.page.getByText('Example Component', { exact: true })
  }

  async goto(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}`)
  }

  async expectLoaded() {
    await expect(this.title).toBeVisible({ timeout: 15_000 })
  }
}
