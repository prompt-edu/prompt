import { Page, Locator, expect } from '@playwright/test'

// /management/courses — the active courses list (cards or table).
export class CoursesPage {
  readonly heading: Locator

  constructor(private readonly page: Page) {
    // level 1 to avoid the sidebar's "COURSES" section label (an <h3>).
    this.heading = page.getByRole('heading', { level: 1, name: 'Courses' })
  }

  async goto() {
    await this.page.goto('/management/courses')
  }

  // Same-document navigation through the shell's sidebar. page.goto() reloads
  // the document, which drops whatever stylesheet a remote injected - so any
  // test about the cascade across remotes has to come back this way.
  async gotoFromSidebar() {
    await this.page.getByTestId('sidebar-home').click()
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible()
  }

  course(name: string): Locator {
    return this.page.getByText(name, { exact: true }).first()
  }

  async expectCourseVisible(name: string) {
    await expect(this.course(name)).toBeVisible()
  }

  card(courseId: string): Locator {
    return this.page.getByTestId(`course-card-${courseId}`)
  }

  // ── Card-view search ─────────────────────────────────────────────────────

  searchInput(): Locator {
    return this.page.getByRole('searchbox', { name: 'Search courses' })
  }

  async search(query: string) {
    await this.searchInput().fill(query)
  }

  noMatchesMessage(): Locator {
    return this.page.getByText('No courses match your search.')
  }

  // Clicks the "Go to course" action within a specific course's card.
  async openCourse(courseId: string) {
    await this.card(courseId)
      .getByRole('button', { name: /go to course/i })
      .click()
  }
}
