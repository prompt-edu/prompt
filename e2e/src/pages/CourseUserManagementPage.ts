import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/user-management — visible to admins and course
// lecturers. Renders two tables (Lecturers / Editors); with no seeded per-course
// Keycloak subgroups both tables render empty, which is the expected soft-fail
// path.
export class CourseUserManagementPage {
  readonly heading: Locator
  readonly addLecturerButton: Locator
  readonly addEditorButton: Locator

  constructor(
    private readonly page: Page,
    private readonly courseId: string,
  ) {
    this.heading = page.getByRole('heading', { level: 1, name: 'User Management' })
    // The card titles are shadcn's <CardTitle> (a plain <div>), not <h*>, so
    // anchor on the per-table "Add Lecturer" / "Add Editor" buttons instead -
    // they are unique and prove both cards rendered.
    this.addLecturerButton = page.getByRole('button', { name: 'Add Lecturer' })
    this.addEditorButton = page.getByRole('button', { name: 'Add Editor' })
  }

  async goto() {
    await this.page.goto(
      `/management/course/${this.courseId}/user-management`,
    )
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible()
    await expect(this.addLecturerButton).toBeVisible()
    await expect(this.addEditorButton).toBeVisible()
  }

  async expectBlocked() {
    // PermissionRestriction swaps in UnauthorizedPage instead of the header.
    await expect(this.heading).toBeHidden()
  }
}
