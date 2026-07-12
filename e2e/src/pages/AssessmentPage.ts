import { Page, Locator, expect } from '@playwright/test'

// Human-readable labels of the five score levels, as rendered by the shared
// ScoreLevelSelector (aria-label "Select <label> score level").
export type ScoreLevelLabel =
  | 'Strongly Disagree'
  | 'Disagree'
  | 'Neutral'
  | 'Agree'
  | 'Strongly Agree'

// /management/course/:courseId/:phaseId — the Assessment remote (Module
// Federation) rendered inside the core shell. Staff see participants/settings/
// grading; students see the evaluation overview and released results.
export class AssessmentPage {
  constructor(private readonly page: Page) {}

  async goto(courseId: string, phaseId: string, subpath = '') {
    await this.page.goto(`/management/course/${courseId}/${phaseId}${subpath}`)
  }

  // ── Overview (phase root) ────────────────────────────────────────────────

  async expectOverviewLoaded() {
    await expect(
      this.page.getByRole('heading', { name: 'Assessment Results & Evaluation' }),
    ).toBeVisible({ timeout: 15_000 })
  }

  // ── Settings ─────────────────────────────────────────────────────────────

  async expectSettingsLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Assessment Settings' })).toBeVisible({
      timeout: 15_000,
    })
  }

  // The "Assessment" card on the settings page (schema select, timeframe,
  // visibility switches, release section, save button).
  settingsCard(): Locator {
    return this.page
      .locator('div.p-6')
      .filter({ has: this.page.getByRole('heading', { name: 'Assessment', exact: true }) })
      .first()
  }

  async createSchema(name: string, description: string) {
    await this.page.getByRole('button', { name: 'Create new assessment schema' }).click()
    const dialog = this.page.getByRole('dialog', { name: 'Create New Assessment Schema' })
    await dialog.getByLabel('Name').fill(name)
    await dialog.getByLabel('Description').fill(description)
    await dialog.getByRole('button', { name: 'Create', exact: true }).click()
    await expect(dialog).toBeHidden()
  }

  // The schema Select sits next to the create-schema button in the
  // "Assessment" settings card.
  async selectSchema(name: string) {
    const card = this.settingsCard()
    await card.getByRole('combobox').first().click()
    await this.page.getByRole('option', { name }).click()
  }

  async saveAssessmentSettings() {
    await this.settingsCard().getByRole('button', { name: 'Save Assessment' }).click()
    // The button disables again once the card has no unsaved changes.
    await expect(
      this.settingsCard().getByRole('button', { name: 'Save Assessment' }),
    ).toBeDisabled()
  }

  async openSchemaDetails() {
    await this.page.getByRole('link', { name: 'Open schema details' }).click()
    await expect(
      this.page.getByRole('heading', { name: 'Categories and competencies' }),
    ).toBeVisible()
  }

  releaseButton(): Locator {
    return this.page.getByRole('button', { name: /Release Results/ })
  }

  // ── Schema configuration (rubric) ────────────────────────────────────────

  async addCategory(name: string, shortName: string, weight = 1) {
    await this.page.getByRole('button', { name: 'Add category' }).click()
    const form = this.page.locator('form', {
      has: this.page.getByRole('button', { name: 'Create Category' }),
    })
    await form.getByLabel('Category Name', { exact: true }).fill(name)
    await form.getByLabel('Short Category Name').fill(shortName)
    await form.getByLabel('Weight').fill(String(weight))
    await form.getByRole('button', { name: 'Create Category' }).click()
    await expect(this.page.getByRole('heading', { name })).toBeVisible()
  }

  async addCompetency(categoryName: string, name: string, shortName: string) {
    await this.page.getByRole('button', { name: `Add Competency to ${categoryName}` }).click()
    const form = this.page.locator('form#competency-form')
    await form.getByLabel('Name', { exact: true }).fill(name)
    await form.getByLabel('Short Name').fill(shortName)
    await form.getByLabel('Description', { exact: true }).fill(`${name} description`)
    await form.getByPlaceholder('Very bad level description').fill('Far below expectations')
    await form.getByPlaceholder('Bad level description', { exact: true }).fill('Below expectations')
    await form.getByPlaceholder('OK level description').fill('Meets expectations')
    await form.getByPlaceholder('Good level description', { exact: true }).fill('Above expectations')
    await form.getByPlaceholder('Very good level description').fill('Far above expectations')
    await form.getByRole('button', { name: 'Create', exact: true }).click()
    await expect(form).toBeHidden()
    await expect(this.page.getByText(name, { exact: true }).first()).toBeVisible()
  }

  // ── Participants ─────────────────────────────────────────────────────────

  async expectParticipantsLoaded() {
    await expect(this.page.getByRole('heading', { name: 'Assessment Participants' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async openParticipant(fullName: string) {
    await this.page.getByRole('row', { name: new RegExp(fullName) }).click()
    await expect(this.page.getByRole('heading', { name: 'Assessment Summary' })).toBeVisible({
      timeout: 15_000,
    })
  }

  // ── Grading (per-participant assessment page) ────────────────────────────

  // Each competency renders as a bordered block containing its name and the
  // five score-level buttons.
  competencyBlock(competencyName: string): Locator {
    return this.page
      .locator('div.border.rounded-md')
      .filter({ hasText: competencyName })
      .filter({ has: this.page.getByRole('button', { name: /score level$/ }) })
  }

  async scoreCompetency(competencyName: string, level: ScoreLevelLabel) {
    const button = this.competencyBlock(competencyName).getByRole('button', {
      name: `Select ${level} score level`,
    })
    await button.click()
    await expect(button).toHaveAttribute('aria-pressed', 'true')
  }

  async fillGeneralRemarks(text: string) {
    const remarks = this.page.getByPlaceholder('What did this person do particularly well?')
    await remarks.fill(text)
    await remarks.blur()
  }

  async selectGradeSuggestion(gradeLabel: string) {
    await this.page.getByText('Select a Grade Suggestion for this Student ...').click()
    await this.page.getByRole('option', { name: gradeLabel }).click()
  }

  async markAssessmentAsFinal() {
    await this.page.getByRole('button', { name: 'Mark Assessment as Final' }).click()
    await this.page.getByRole('button', { name: 'Yes, Mark as Final' }).click()
    await expect(this.page.getByRole('button', { name: 'Unmark as Final' })).toBeVisible()
  }

  // ── Student results ──────────────────────────────────────────────────────

  async gotoResults(courseId: string, phaseId: string) {
    await this.goto(courseId, phaseId, '/results')
    await expect(this.page.getByRole('heading', { name: 'Assessment Results' })).toBeVisible({
      timeout: 15_000,
    })
  }

  async expectResultsNotReleased() {
    await expect(
      this.page.getByText('Assessment results have not been released yet.'),
    ).toBeVisible()
  }

  // ── Student self evaluation ──────────────────────────────────────────────

  async openSelfEvaluation() {
    // The overview groups the self evaluation under a "Self Evaluation" section
    // heading; the clickable target row inside it is labeled "Evaluate yourself".
    await this.page.getByText('Evaluate yourself', { exact: true }).click()
    await expect(
      this.page.getByText('Please fill out the self-evaluation below', { exact: false }),
    ).toBeVisible()
  }

  async markEvaluationAsFinal() {
    await this.page.getByRole('button', { name: 'Mark as Final' }).click()
    await this.page.getByRole('button', { name: 'Yes, Mark as Final' }).click()
    await expect(this.page.getByRole('button', { name: 'Unmark as Final' })).toBeVisible()
  }

  // ── Print report ─────────────────────────────────────────────────────────

  // The hidden report (`.print-report`, `hidden print:block`) that becomes the
  // printed page under print media. On-screen it has no box.
  printReport(): Locator {
    return this.page.locator('.print-report')
  }

  // The core shell's sidebar wrapper: shadcn renders a gap/spacer div plus a
  // fixed sidebar under one `data-side` wrapper, the only direct child of the
  // SidebarProvider wrapper carrying that attribute. The print fix hides it so
  // the report is not pushed to the right half of the page.
  private sidebarWrapper(): Locator {
    return this.page.locator('.group\\/sidebar-wrapper > [data-side]')
  }

  // Under print media the sidebar must be gone and the report must fill the
  // page (regression guard for the "renders only on the right half" bug). Resets
  // media emulation before returning.
  async expectPrintReportFillsPage() {
    await this.page.emulateMedia({ media: 'print' })
    try {
      const report = this.printReport()
      await expect(report).toBeVisible()
      await expect(this.sidebarWrapper()).toBeHidden()

      const viewport = this.page.viewportSize()
      const box = await report.boundingBox()
      expect(viewport, 'viewport size').not.toBeNull()
      expect(box, 'print report bounding box').not.toBeNull()
      // The sidebar-hidden assertion above is the real regression guard; this
      // just confirms the report is not squeezed to the right half. Pre-fix the
      // reserved sidebar spacer (~271px expanded) pushed x right and shrank the
      // report; full-width means a near-left start spanning most of the page.
      expect(box!.x).toBeLessThan(60)
      expect(box!.width).toBeGreaterThan(viewport!.width * 0.9)
    } finally {
      await this.page.emulateMedia({ media: null })
    }
  }
}
