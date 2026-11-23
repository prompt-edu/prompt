import { Page } from '@playwright/test'
import { pickDateRange, pickOrg, waitForCourseId } from './CourseCreationPage'

export interface CopyCourseInput {
  templateName: string
  name: string
  semesterTag: string
  courseType?: string
  // Display name of the org to create the course in, for a user with org roles
  org?: string
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

    if (input.courseType) {
      // The course type comes before the org select
      await form.getByRole('combobox').first().click()
      await this.page.getByRole('option', { name: input.courseType }).click()
    }

    // Copy form requires a date range with to > from.
    await pickDateRange(this.page)
    if (input.org) {
      await pickOrg(this.page, form, input.org)
    }

    await form.getByRole('button', { name: 'Continue' }).click()

    // Warning step: the proceed button's label is dynamic, so target the testid.
    await this.page.getByTestId('copy-course-confirm').click()

    return await waitForCourseId(this.page)
  }
}
