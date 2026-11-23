import { Page, Locator, expect } from '@playwright/test'

export interface CreateCourseInput {
  name: string
  semesterTag: string
  // Display name of the course type. 'Lecture' has no fixed ECTS, so the ECTS
  // field stays editable and the test exercises it.
  courseType?: string
  ects?: number
  shortDescription?: string
  // Display name of the org to create the course in, for a user with org roles
  org?: string
}

// The two-page "Add a New Course" wizard, reached from the course-switch
// sidebar's "+" button.
export class CourseCreationPage {
  constructor(private readonly page: Page) {}

  private dialog(): Locator {
    return this.page.getByRole('dialog', { name: 'Add a New Course' })
  }

  async openCreateDialog(): Promise<Locator> {
    await this.page.goto('/management/courses')
    await this.page.getByTestId('add-course-button').click()
    await this.page.getByRole('button', { name: /Create New Course/ }).click()
    const dialog = this.dialog()
    await expect(dialog).toBeVisible()
    return dialog
  }

  async create(input: CreateCourseInput): Promise<string> {
    const dialog = await this.openCreateDialog()
    await dialog.getByPlaceholder('Enter a course name').fill(input.name)

    // Select the course type first: a fixed-ECTS type disables the ECTS input.
    await dialog.getByRole('combobox').filter({ hasText: 'Select a course type' }).click()
    await this.page.getByRole('option', { name: input.courseType ?? 'Lecture' }).click()

    await pickDateRange(this.page)

    const ects = dialog.getByPlaceholder('Enter ECTS')
    if (await ects.isEnabled()) {
      await ects.fill(String(input.ects ?? 6))
    }
    await dialog.getByPlaceholder('Enter a semester tag').fill(input.semesterTag)
    await dialog
      .getByPlaceholder('One sentence summary')
      .fill(input.shortDescription ?? 'created by e2e')
    if (input.org) {
      await pickOrg(this.page, dialog, input.org)
    }

    await dialog.getByRole('button', { name: 'Next', exact: true }).click()

    // Appearance page: pick a color + icon, then submit.
    await dialog.getByRole('combobox').filter({ hasText: 'Select a color' }).click()
    await this.page.getByRole('option', { name: 'blue' }).click()
    await dialog.getByRole('combobox').filter({ hasText: 'Select an icon' }).click()
    await this.page.getByRole('option', { name: 'Monitor' }).click()
    await dialog.getByRole('button', { name: 'Add Course', exact: true }).click()

    return await waitForCourseId(this.page)
  }
}

// Drives the shared DatePickerWithRange (its "Open calendar" button opens a
// react-day-picker range calendar on the current month). Days 10 and 20 are
// always in-month (never a [data-outside] cell), and 20 > 10 satisfies the copy
// form's to > from rule.
export async function pickDateRange(page: Page): Promise<void> {
  await page.getByRole('button', { name: 'Open calendar' }).click()
  const day = (n: number) =>
    page.locator('td[data-day]:not([data-outside]) button', {
      hasText: new RegExp(`^${n}$`),
    })
  await day(10).click()
  await day(20).click()
  await page.keyboard.press('Escape')
}

// The org select of the create and copy forms
export async function pickOrg(page: Page, form: Locator, org: string): Promise<void> {
  await form.getByRole('combobox').filter({ hasText: 'Choose an org' }).click()
  await page.getByRole('option', { name: org, exact: true }).click()
}

// After a successful create/copy the app navigates to the new course overview.
export async function waitForCourseId(page: Page): Promise<string> {
  await page.waitForURL(/\/management\/course\/[0-9a-f-]{36}/)
  const match = page.url().match(/\/management\/course\/([0-9a-f-]{36})/)
  if (!match) throw new Error(`could not parse course id from ${page.url()}`)
  return match[1]
}
