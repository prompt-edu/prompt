import { Page, Locator, expect } from '@playwright/test'

// /management/course/:courseId/:phaseId — the Self Team Allocation remote
// (Module Federation) rendered inside the core shell. Students see the
// create/join team UI; staff see the read-only team overview.
export class SelfTeamAllocationPage {
  readonly heading: Locator

  constructor(private readonly page: Page) {
    this.heading = this.page.getByRole('heading', { level: 1, name: 'Team Allocation' })
  }

  async goto(courseId: string, phaseId: string) {
    await this.page.goto(`/management/course/${courseId}/${phaseId}`)
  }

  async expectLoaded() {
    await expect(this.heading).toBeVisible({ timeout: 15_000 })
  }

  teamCard(name: string): Locator {
    return this.page.getByTestId(`team-card-${name}`)
  }

  async createTeam(name: string) {
    await this.page.getByRole('button', { name: 'Create new team' }).click()
    await this.page.getByPlaceholder('Enter team name').fill(name)
    await this.page.getByRole('button', { name: 'Create team', exact: true }).click()
    await expect(this.teamCard(name)).toBeVisible()
  }

  // Creating a team does NOT assign the creator; joining is a separate action.
  async joinTeam(name: string) {
    await this.teamCard(name).getByRole('button', { name: 'Join Team' }).click()
    await expect(this.teamCard(name).getByText('You')).toBeVisible()
  }

  async leaveTeam(name: string) {
    await this.teamCard(name).getByRole('button', { name: 'Leave Team' }).click()
    await expect(this.teamCard(name).getByText('You')).toBeHidden()
  }

  async expectMemberCount(name: string, count: number) {
    await expect(this.teamCard(name).getByText(`${count}/3 Members`)).toBeVisible()
  }

  async expectMemberVisible(name: string, memberName: string) {
    await expect(this.teamCard(name).getByText(memberName)).toBeVisible()
  }
}
