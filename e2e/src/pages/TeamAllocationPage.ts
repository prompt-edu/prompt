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
      this.page.getByRole('heading', { level: 1, name: 'Settings', exact: true }),
    ).toBeVisible({ timeout: 15_000 })
  }

  async expectPreviousPhaseDataCheck() {
    await expect(this.page.getByText('Previous Phase Data', { exact: true })).toBeVisible({
      timeout: 15_000,
    })
    await expect(this.page.getByText(/of \d+ checks complete/)).toBeVisible({ timeout: 15_000 })
  }

  launchDialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Launch TEASE' })
  }

  // Expects both launch warnings: the phase has no previous phase data and no survey submission.
  async expectLaunchDialogWarnings() {
    await this.page.getByRole('button', { name: 'Launch Tease to Matchmake' }).click()
    const dialog = this.launchDialog()
    await expect(dialog).toBeVisible()
    const launch = dialog.getByRole('link', { name: /^Launch TEASE/ })
    await expect(launch).toHaveAttribute('target', '_blank')
    await expect(launch).toHaveAttribute('href', /\/tease\?coursePhaseId=/)
    await expect(dialog.getByText(/^Data from previous phases is missing:/)).toBeVisible({
      timeout: 15_000,
    })
    await expect(dialog.getByRole('link', { name: 'Open the settings' })).toHaveAttribute(
      'href',
      /\/settings$/,
    )
    await expect(dialog.getByText(/students have not submitted the survey/)).toBeVisible()
    await expect(launch).toHaveText(/Launch TEASE anyway/)
  }

  async followNotSubmittedLink() {
    await this.launchDialog().getByRole('link', { name: 'Show these students' }).click()
    await expect(this.page).toHaveURL(/\/participants\?filters=/)
  }

  async expectOnlyNotSubmittedParticipants(lastName: string) {
    const table = this.page.locator('#table-view')
    await expect(table.getByRole('columnheader', { name: 'Survey' })).toBeVisible({
      timeout: 15_000,
    })
    const row = table.getByRole('row').filter({ hasText: lastName })
    await expect(row.getByText('Not submitted', { exact: true })).toBeVisible()
    await expect(table.getByText('Submitted', { exact: true })).toHaveCount(0)
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
