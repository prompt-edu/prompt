import { Page } from '@playwright/test'
import { waitForCourseId } from './CourseCreationPage'

export interface CopyCourseInput {
  templateName: string
  name: string
  semesterTag: string
}

// The "Use Template" flow: choice dialog -> template picker -> copy form ->
// compatibility warning step -> new course.
export class CopyCoursePage {
  constructor(private readonly page: Page) {}

  async copyFromTemplate(input: CopyCourseInput): Promise<string> {
    await this.page.goto('/management/courses')
    await this.page.getByTestId('add-course-button').click()
    await this.page.getByRole('button', { name: /Use Template/ }).click()

    const chooser = this.page.getByRole('dialog', { name: 'Choose a Template' })
    await chooser.getByText(input.templateName, { exact: true }).click()

    const form = this.page
      .getByRole('dialog')
      .filter({ hasText: 'Create Course from Template' })
    await form.getByPlaceholder('Enter course name').fill(input.name)
    await form.getByPlaceholder('Enter semester tag').fill(input.semesterTag)

    // Copy form requires a date range with to > from.
    await this.page.locator('#date').click()
    const day = (n: number) =>
      this.page.locator('.rdp button[name="day"]:not(.day-outside)', {
        hasText: new RegExp(`^${n}$`),
      })
    await day(10).click()
    await day(20).click()
    await this.page.keyboard.press('Escape')

    await form.getByRole('button', { name: 'Continue' }).click()

    // Warning step: the proceed button's label is dynamic, so target the testid.
    await this.page.getByTestId('copy-course-confirm').click()

    return await waitForCourseId(this.page)
  }
}
