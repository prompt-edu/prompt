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
    return this.page.getByRole('button', { name })
  }

  // The enrolled phases show up under the "Course Phases" sidebar group; a
  // student only sees the phases their participation makes active.
  async expectPhaseListed(name: string) {
    await expect(this.page.getByText('Course Phases')).toBeVisible()
    await expect(this.phaseItem(name)).toBeVisible()
  }
}
