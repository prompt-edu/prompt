import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/user-management — visible to admins and course
// lecturers. Renders two tables (Lecturers / Editors); with no seeded per-course
// Keycloak subgroups both tables render empty, which is the expected soft-fail
// path.
export type StaffGroup = 'Lecturer' | 'Editor'

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

  // The remove button carries the username and the group in its label, so it
  // identifies a member row of one specific table.
  removeButton(group: StaffGroup, username: string): Locator {
    return this.page.getByRole('button', { name: `Remove ${username} from ${group}` })
  }

  async addMember(group: StaffGroup, username: string) {
    await this.page.getByRole('button', { name: `Add ${group}` }).click()
    const dialog = this.page.getByRole('dialog')
    await dialog.getByRole('textbox').fill(username)
    await dialog
      .getByRole('listitem')
      .filter({ hasText: username })
      .getByRole('button', { name: 'Add' })
      .click()
    await expect(dialog).toBeHidden()
  }

  async removeMember(group: StaffGroup, username: string) {
    await this.removeButton(group, username).click()
    await this.page.getByRole('alertdialog').getByRole('button', { name: 'Delete' }).click()
  }

  async expectBlocked() {
    // PermissionRestriction swaps in UnauthorizedPage instead of the header.
    await expect(this.heading).toBeHidden()
  }
}
