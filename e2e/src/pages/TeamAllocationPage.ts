import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId — the Team Allocation remote (Module
// Federation) rendered inside the core shell. Students land on the survey;
// staff get the settings, allocations, and participants views.
export class TeamAllocationPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string, subpath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subpath}`)
  }

  async expectAllocationsLoaded() {
    await expect(
      this.page.getByRole('heading', { level: 1, name: 'Team Allocations' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async expectSurveyLoaded() {
    await expect(
      this.page.getByRole('heading', { level: 1, name: 'Team Allocation Survey' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async expectSettingsLoaded() {
    await expect(
      this.page.getByRole('heading', { level: 1, name: 'Survey Settings' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async createTeam(name: string) {
    await this.page.getByPlaceholder('Enter new team name').fill(name)
    await this.page.getByRole('button', { name: 'Add Team' }).click()
    await expect(this.page.getByText(name, { exact: true })).toBeVisible({ timeout: 15_000 })
  }

  teamCardHeading(name: string): Locator {
    return this.page.getByRole('heading', { level: 3, name, exact: true })
  }

  async expectAllocatedMember(teamName: string, memberFullName: string) {
    await expect(this.teamCardHeading(teamName)).toBeVisible({ timeout: 15_000 })
    await expect(this.page.getByText(memberFullName)).toBeVisible()
  }

  async expectParticipantAllocatedTeam(lastName: string, teamName: string) {
    const table = this.page.locator('#table-view')
    await expect(
      table.getByRole('heading', { level: 1, name: 'Team Allocation Participants' }),
    ).toBeVisible({ timeout: 15_000 })
    await expect(table.getByText(lastName).first()).toBeVisible()
    await expect(table.getByText(teamName).first()).toBeVisible()
  }
}
