import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId — a course's overview, the landing view a
// student sees for a course they are enrolled in. The course name is a
// CardTitle (not a role heading); the enrolled phases are listed in the left
// "Course Phases" sidebar group.
export class CourseOverviewPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string) {
    await this.page.goto(`/management/course/${courseId}`)
  }

  courseName(name: string): Locator {
    return this.page.getByText(name, { exact: true }).first()
  }

  async expectLoaded(courseName: string) {
    await expect(this.courseName(courseName)).toBeVisible({ timeout: 15_000 })
  }

  phaseItem(name: string): Locator {
    return this.page.getByRole('button', { name, exact: true })
  }

  // The enrolled phases show up under the "Course Phases" sidebar group; a
  // student only sees the phases their participation makes active.
  async expectPhaseListed(name: string) {
    // Exact match: the sidebar's empty-state copy ("No course phases yet.")
    // also contains the substring "course phases", so a loose getByText would
    // match two elements and trip Playwright strict mode.
    await expect(this.page.getByText('Course Phases', { exact: true })).toBeVisible()
    await expect(this.phaseItem(name)).toBeVisible()
  }

  // The phase sidebars are Module Federation remotes behind React.lazy, so an
  // "entry is absent" assertion has to outlive their loading: a pending remote
  // renders a disabled "Loading..." placeholder, and any rendered menu item
  // CSS-hides the empty state. Waiting for both rules out a pass that only
  // means the remote had not loaded yet.
  async expectNoPhaseListed(name: string) {
    await expect(this.page.getByText('Course Phases', { exact: true })).toBeVisible()
    await expect(this.page.getByRole('button', { name: 'Loading...', exact: true })).toHaveCount(0)
    await expect(this.page.getByText('No course phases yet.')).toBeVisible()
    await expect(this.phaseItem(name)).toHaveCount(0)
  }
}
